package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
)

type Configuration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DB       string `json:"db"`
	User     string `json:"user"`
	Password string `json:"password"`
	Schema   Schema `json:"schema"`
}

type Schema struct {
	ScalarTypes map[string]ScalarType `json:"scalar_types"`
	ObjectTypes map[string]ObjectType `json:"object_types"`
	Collections []Collection          `json:"collections"`
	Functions   []interface{}         `json:"functions"`
	Procedures  []interface{}         `json:"procedures"`
}

type ScalarType struct {
	AggregateFunctions  map[string]AggregateFunction `json:"aggregate_functions"`
	ComparisonOperators map[string]Operator          `json:"comparison_operators"`
	UpdateOperators     map[string]interface{}       `json:"update_operators"`
}

type AggregateFunction struct {
	ResultType DataType `json:"result_type"`
}

type Operator struct {
	ArgumentType DataType `json:"argument_type"`
}

type DataType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ObjectType struct {
	Description string           `json:"description"`
	Fields      map[string]Field `json:"fields"`
}

type Field struct {
	Description string              `json:"description"`
	Arguments   map[string]Argument `json:"arguments"`
	Type        DataType            `json:"type"`
}

type Argument struct {
	Description string   `json:"description"`
	Type        DataType `json:"type"`
}

type Collection struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Arguments             map[string]interface{} `json:"arguments"`
	Type                  string                 `json:"type"`
	InsertableColumns     []interface{}          `json:"insertable_columns"`
	UpdatableColumns      []interface{}          `json:"updatable_columns"`
	Deletable             bool                   `json:"deletable"`
	UniquenessConstraints map[string]interface{} `json:"uniqueness_constraints"`
	ForeignKeys           map[string]ForeignKey  `json:"foreign_keys"`
}

type ForeignKey struct {
	ColumnMapping     map[string]string `json:"column_mapping"`
	ForeignCollection string            `json:"foreign_collection"`
}

type State struct {
	Database  *sql.DB
	Telemetry *connector.TelemetryState
}

type Connector struct{}

func (mc *Connector) Query(ctx context.Context, configuration *Configuration, state *State, request *schema.QueryRequest) (schema.QueryResponse, error) {

	variableSets := request.Variables
	if variableSets == nil {
		variableSets = []schema.QueryRequestVariablesElem{make(map[string]any)}
	}
	rowSets := make([]schema.RowSet, 0, len(variableSets))

	sql := getFetchQuery(request)

	rows, err := state.Database.Query(sql)
	if err != nil {
		fmt.Print("Database query failed!")
	}
	defer rows.Close()

	var rowSet schema.RowSet
	rowSet.Rows = []map[string]any{}

	cols, err := rows.Columns()
	if err != nil {
		return schema.QueryResponse{}, err
	}

	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return schema.QueryResponse{}, err
		}

		rowMap := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*any)
			switch v := (*val).(type) {
			case []byte:
				rowMap[colName] = string(v)
			default:
				rowMap[colName] = *val
			}
		}

		rowSet.Rows = append(rowSet.Rows, rowMap)
	}

	if err := rows.Err(); err != nil {
		return schema.QueryResponse{}, err
	}

	rowSets = append(rowSets, rowSet)
	return rowSets, nil
}

func getFetchQuery(request *schema.QueryRequest) string {
	var fields []string

	for _, field := range request.Query.Fields {
		if fieldType, ok := field["type"].(schema.FieldType); ok && fieldType == schema.FieldTypeColumn {
			if columnName, ok := field["column"].(string); ok {
				fields = append(fields, columnName)
			}
		}
	}

	selectClause := "SELECT " + strings.Join(fields, ", ") + " FROM " + request.Collection

	limitClause := ""
	if request.Query.Limit != nil {
		limitClause = fmt.Sprintf("LIMIT %d", *request.Query.Limit)
	}

	offsetClause := ""
	if request.Query.Offset != nil {
		offsetClause = fmt.Sprintf("OFFSET %d", *request.Query.Offset)
	}

	orderByClause := ""
	if request.Query.OrderBy != nil && len(request.Query.OrderBy.Elements) > 0 {
		var orderByElements []string
		for _, element := range request.Query.OrderBy.Elements {
			if targetName, ok := element.Target["name"].(string); ok {
				orderByElements = append(orderByElements, fmt.Sprintf("%s %s", targetName, element.OrderDirection))
			}
		}
		orderByClause = "ORDER BY " + strings.Join(orderByElements, ", ")
	}

	whereClause := ""
	if request.Query.Predicate != nil {
		whereClause = "WHERE " + visitExpression(request.Query.Predicate)
	}

	sql := fmt.Sprintf("%s %s %s %s %s", selectClause, orderByClause, whereClause, limitClause, offsetClause)

	return sql
}

func visitExpression(expression schema.Expression) string {
	expressionType, err := expression.Type()
	if err != nil {
		fmt.Print("Invalid expression type in the predicate")
	}
	switch expressionType {
	case schema.ExpressionTypeAnd:
		return visitLogicalExpression(expression, "AND")
	case schema.ExpressionTypeOr:
		return visitLogicalExpression(expression, "OR")
	case schema.ExpressionTypeNot:
		return "NOT " + visitExpression(expression)
	case schema.ExpressionTypeUnaryComparisonOperator:
		return visitUnaryComparison(expression)
	case schema.ExpressionTypeBinaryComparisonOperator:
		return visitBinaryComparison(expression)
	default:
		return ""
	}
}

func visitLogicalExpression(expression schema.Expression, operator string) string {
	var clauses []string
	subExpressions := expression["expressions"].([]schema.Expression)
	for _, subExpression := range subExpressions {
		clauses = append(clauses, visitExpression(subExpression))
	}
	return "(" + strings.Join(clauses, " "+operator+" ") + ")"
}

func visitUnaryComparison(expression schema.Expression) string {
	targetName := expression["column"].(schema.ComparisonTarget).Name
	operator := expression["operator"]
	return fmt.Sprintf("%s %s", targetName, operator)
}

func visitBinaryComparison(expression schema.Expression) string {
	targetName := expression["column"].(schema.ComparisonTarget).Name
	operator := expression["operator"]
	var clause string
	switch operator {
	case "_lte":
		clause = "<="
	case "_in":
		clause = "IN"
	case "_like":
		clause = "LIKE"
	case "_eq":
		clause = "="
	default:
		fmt.Print("Invalid comparison operator!")
	}
	value := getComparisonValue(expression)
	return fmt.Sprintf("%s %s %s", targetName, clause, value)
}

func getComparisonValue(expression map[string]interface{}) string {
	comparisonValue := expression["value"].(schema.ComparisonValue)["value"]

	switch value := comparisonValue.(type) {
	case []interface{}:
		formattedValues := make([]string, len(value))
		for i, item := range value {
			formattedValues[i] = fmt.Sprintf("\"%v\"", item)
		}
		return fmt.Sprintf("(%s)", strings.Join(formattedValues, ", "))
	default:
		return fmt.Sprintf("'%v'", value)
	}
}

func (mc *Connector) GetCapabilities(configuration *Configuration) schema.CapabilitiesResponseMarshaler {
	return nil
}

func (mc *Connector) GetSchema(ctx context.Context, configuration *Configuration, state *State) (schema.SchemaResponseMarshaler, error) {
	return nil, nil
}

func (mc *Connector) HealthCheck(ctx context.Context, configuration *Configuration, state *State) error {
	return nil
}

func (mc *Connector) Mutation(ctx context.Context, configuration *Configuration, state *State, request *schema.MutationRequest) (*schema.MutationResponse, error) {
	return nil, nil
}

func (mc *Connector) MutationExplain(ctx context.Context, configuration *Configuration, state *State, request *schema.MutationRequest) (*schema.ExplainResponse, error) {
	return &schema.ExplainResponse{
		Details: schema.ExplainResponseDetails{},
	}, nil
}

func (mc *Connector) ParseConfiguration(ctx context.Context, rawConfiguration string) (*Configuration, error) {
	data, err := readConfigFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
		return nil, nil
	}

	var config Configuration

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error unmarshaling config file: %v\n", err)
		return nil, nil
	}

	return &config, nil
}

func (mc *Connector) QueryExplain(ctx context.Context, configuration *Configuration, state *State, request *schema.QueryRequest) (*schema.ExplainResponse, error) {
	return nil, schema.NotSupportedError("query explain has not been supported yet", nil)
}

func (mc *Connector) TryInitState(ctx context.Context, configuration *Configuration, metrics *connector.TelemetryState) (*State, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", configuration.User, configuration.Password, configuration.Host, configuration.Port, configuration.DB)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
		return nil, err
	} else {
		fmt.Println("Database connected successfully")
	}

	return &State{
		Database:  db,
		Telemetry: metrics,
	}, nil
}

func readConfigFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}
