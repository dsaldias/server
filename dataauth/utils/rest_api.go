package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi"
)

// RestToGraphQlHandler convierte requests REST a GraphQL ejecutándolos
// contra el mismo schema de gqlgen sin duplicar lógica de negocio.
//
// ─── Uso para un developer REST ──────────────────────────────────────────────
//
// GET  /rest/query/mis_tickets                       → todos los campos automáticamente
// GET  /rest/query/mis_tickets?_fields=id,estado     → solo campos pedidos
//
// POST /rest/mutation/create_ticket
//
//	{ "input": { "problema": "Holis" } }          → sin necesidad de _types
func RestToGraphQlHandler(schema graphql.ExecutableSchema) http.HandlerFunc {
	gqlHandler := handler.NewDefaultServer(schema)
	fieldCache := &sync.Map{}

	return func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		operationName := rctx.URLParam("operationName")
		if operationName == "" {
			writeJSONError(w, "operationName es requerido", http.StatusBadRequest)
			return
		}

		variables, fields, err := extractParams(r)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("parámetros inválidos: %s", err), http.StatusBadRequest)
			return
		}

		if fields == "" {
			fields = getCachedFields(r.Context(), gqlHandler, fieldCache, operationName, isMutation(r))
		}

		// Construir query con valores literales incrustados (sin variables GraphQL)
		// Igual que escribirlo a mano en el playground — sin necesidad de tipos.
		gqlQuery := buildQueryWithLiterals(operationName, variables, fields, isMutation(r))

		// Ejecutar sin variables — todo está incrustado en el query string
		result, err := executeAgainstSchema(r.Context(), gqlHandler, gqlQuery, nil)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, hasErrors := result["errors"]; hasErrors {
			w.WriteHeader(http.StatusBadRequest)
		}
		json.NewEncoder(w).Encode(result)
	}
}

// ─── buildQueryWithLiterals ──────────────────────────────────────────────────
// Genera un query GraphQL con los valores incrustados directamente,
// igual que escribirlo a mano en el playground.
//
// Input:  operationName="create_ticket", variables={"input":{"problema":"holis"}}
// Output: mutation RestProxy { create_ticket(input: {problema: "holis"}) { id } }
func buildQueryWithLiterals(operationName string, variables map[string]interface{}, fields string, mutation bool) string {
	opType := "query"
	if mutation {
		opType = "mutation"
	}

	selectionSet := buildSelectionSet(fields)

	if len(variables) == 0 {
		return fmt.Sprintf(`%s RestProxy { %s %s }`, opType, operationName, selectionSet)
	}

	args := make([]string, 0, len(variables))
	for k, v := range variables {
		args = append(args, fmt.Sprintf("%s: %s", k, toGraphQLLiteral(v)))
	}

	return fmt.Sprintf(
		`%s RestProxy { %s(%s) %s }`,
		opType,
		operationName,
		strings.Join(args, ", "),
		selectionSet,
	)
}

// ─── toGraphQLLiteral ────────────────────────────────────────────────────────
// Convierte un valor Go a su representación literal en GraphQL.
//
//	"hola"           → "hola"
//	42               → 42
//	true             → true
//	{"a":"b"}        → {a: "b"}
//	["x","y"]        → ["x", "y"]
func toGraphQLLiteral(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"

	case float64:
		// JSON decodifica todos los números como float64
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)

	case string:
		// Escapar caracteres especiales dentro del string
		escaped := strings.ReplaceAll(val, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		escaped = strings.ReplaceAll(escaped, "\r", `\r`)
		escaped = strings.ReplaceAll(escaped, "\t", `\t`)
		return fmt.Sprintf(`"%s"`, escaped)

	case map[string]interface{}:
		// Objeto → { clave: valor, ... }
		parts := make([]string, 0, len(val))
		for k, v2 := range val {
			parts = append(parts, fmt.Sprintf("%s: %s", k, toGraphQLLiteral(v2)))
		}
		return "{ " + strings.Join(parts, ", ") + " }"

	case []interface{}:
		// Array → [valor, valor, ...]
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, toGraphQLLiteral(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"

	default:
		// Fallback: serializar como JSON y usar como string
		b, err := json.Marshal(v)
		if err != nil {
			return `""`
		}
		return string(b)
	}
}

// ─── getCachedFields ─────────────────────────────────────────────────────────

func getCachedFields(ctx context.Context, srv *handler.Server, cache *sync.Map, operationName string, mutation bool) string {
	if cached, ok := cache.Load(operationName); ok {
		return cached.(string)
	}
	fields := introspectReturnFields(ctx, srv, operationName, mutation)
	if fields != "" {
		cache.Store(operationName, fields)
	}
	return fields
}

// ─── introspectReturnFields ──────────────────────────────────────────────────

func introspectReturnFields(ctx context.Context, srv *handler.Server, operationName string, mutation bool) string {
	rootType := "Query"
	if mutation {
		rootType = "Mutation"
	}

	typeQuery := fmt.Sprintf(`{
		__type(name: "%s") {
			fields {
				name
				type {
					name kind
					ofType { name kind ofType { name kind ofType { name kind } } }
				}
			}
		}
	}`, rootType)

	result, err := executeAgainstSchema(ctx, srv, typeQuery, nil)
	if err != nil {
		return ""
	}

	returnTypeName := extractReturnTypeName(result, operationName)
	if returnTypeName == "" {
		return ""
	}

	return introspectTypeFields(ctx, srv, returnTypeName)
}

// ─── extractReturnTypeName ───────────────────────────────────────────────────

func extractReturnTypeName(result map[string]interface{}, operationName string) string {
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return ""
	}
	typeInfo, ok := data["__type"].(map[string]interface{})
	if !ok {
		return ""
	}
	fields, ok := typeInfo["fields"].([]interface{})
	if !ok {
		return ""
	}
	for _, f := range fields {
		field, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		if field["name"] != operationName {
			continue
		}
		typeObj, ok := field["type"].(map[string]interface{})
		if !ok {
			continue
		}
		return unwrapTypeName(typeObj)
	}
	return ""
}

func unwrapTypeName(t map[string]interface{}) string {
	kind, _ := t["kind"].(string)
	name, _ := t["name"].(string)
	if kind == "OBJECT" || kind == "SCALAR" || kind == "ENUM" || kind == "INTERFACE" {
		return name
	}
	if ofType, ok := t["ofType"].(map[string]interface{}); ok {
		return unwrapTypeName(ofType)
	}
	return name
}

// ─── introspectTypeFields ────────────────────────────────────────────────────

func introspectTypeFields(ctx context.Context, srv *handler.Server, typeName string) string {
	query := fmt.Sprintf(`{
		__type(name: "%s") {
			fields {
				name
				type { kind ofType { kind } }
			}
		}
	}`, typeName)

	result, err := executeAgainstSchema(ctx, srv, query, nil)
	if err != nil {
		return ""
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return ""
	}
	typeInfo, ok := data["__type"].(map[string]interface{})
	if !ok {
		return ""
	}
	fields, ok := typeInfo["fields"].([]interface{})
	if !ok {
		return ""
	}

	var names []string
	for _, f := range fields {
		field, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := field["name"].(string)
		typeObj, ok := field["type"].(map[string]interface{})
		if !ok {
			continue
		}
		if isScalarOrEnum(typeObj) {
			names = append(names, name)
		}
	}
	return strings.Join(names, " ")
}

func isScalarOrEnum(t map[string]interface{}) bool {
	kind, _ := t["kind"].(string)
	switch kind {
	case "SCALAR", "ENUM":
		return true
	case "NON_NULL":
		if ofType, ok := t["ofType"].(map[string]interface{}); ok {
			return isScalarOrEnum(ofType)
		}
	}
	return false
}

// ─── isMutation ──────────────────────────────────────────────────────────────

func isMutation(r *http.Request) bool {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// ─── extractParams ───────────────────────────────────────────────────────────

func extractParams(r *http.Request) (vars map[string]interface{}, fields string, err error) {
	vars = make(map[string]interface{})

	if r.Method == http.MethodGet {
		for k, v := range r.URL.Query() {
			if k == "_fields" {
				fields = v[0]
			} else if len(v) == 1 {
				vars[k] = v[0]
			} else {
				anySlice := make([]interface{}, len(v))
				for i, s := range v {
					anySlice[i] = s
				}
				vars[k] = anySlice
			}
		}
		return
	}

	if r.Body == nil || r.ContentLength == 0 {
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&vars); err != nil {
		err = fmt.Errorf("body JSON inválido: %w", err)
		return
	}

	if f, ok := vars["_fields"]; ok {
		if fStr, ok := f.(string); ok {
			fields = fStr
		}
		delete(vars, "_fields")
	}

	return
}

// ─── buildSelectionSet ───────────────────────────────────────────────────────

func buildSelectionSet(fields string) string {
	if fields == "" {
		return "{ __typename }"
	}
	if strings.Contains(fields, "{") {
		trimmed := strings.TrimSpace(fields)
		if !strings.HasPrefix(trimmed, "{") {
			return "{ " + trimmed + " }"
		}
		return trimmed
	}
	if !strings.Contains(fields, ",") {
		return "{ " + strings.TrimSpace(fields) + " }"
	}

	var sb strings.Builder
	sb.WriteString("{ ")
	depth := 0
	current := strings.Builder{}

	for _, ch := range fields {
		switch ch {
		case '{':
			depth++
			current.WriteString(" { ")
		case '}':
			depth--
			current.WriteString(" } ")
		case ',':
			if depth == 0 {
				if token := strings.TrimSpace(current.String()); token != "" {
					sb.WriteString(token + " ")
				}
				current.Reset()
			} else {
				current.WriteRune(' ')
			}
		default:
			current.WriteRune(ch)
		}
	}
	if token := strings.TrimSpace(current.String()); token != "" {
		sb.WriteString(token + " ")
	}
	sb.WriteString("}")
	return sb.String()
}

// ─── executeAgainstSchema ────────────────────────────────────────────────────

type internalRequestKey struct{}

type gqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

var reqBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var respBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func executeAgainstSchema(
	ctx context.Context,
	srv *handler.Server,
	query string,
	variables map[string]interface{},
) (map[string]interface{}, error) {

	reqPayload := gqlRequest{
		Query:     query,
		Variables: variables,
	}

	reqBuf := reqBufPool.Get().(*bytes.Buffer)
	reqBuf.Reset()
	defer reqBufPool.Put(reqBuf)

	if err := json.NewEncoder(reqBuf).Encode(reqPayload); err != nil {
		return nil, fmt.Errorf("error serializando request GraphQL: %w", err)
	}

	internalCtx := context.WithValue(ctx, internalRequestKey{}, true)
	req, err := http.NewRequestWithContext(internalCtx, http.MethodPost, "/graphql", bytes.NewReader(reqBuf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("error creando request interno: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := newResponseRecorder()
	srv.ServeHTTP(rec, req)

	var result map[string]interface{}
	if err := json.Unmarshal(rec.buf.Bytes(), &result); err != nil {
		rec.reset()
		return nil, fmt.Errorf("error parseando respuesta GraphQL: %w", err)
	}
	rec.reset()
	return result, nil
}

// ─── responseRecorder ────────────────────────────────────────────────────────

type responseRecorder struct {
	headers http.Header
	buf     *bytes.Buffer
	code    int
}

func newResponseRecorder() *responseRecorder {
	buf := respBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return &responseRecorder{
		headers: make(http.Header),
		buf:     buf,
		code:    http.StatusOK,
	}
}

func (r *responseRecorder) Header() http.Header         { return r.headers }
func (r *responseRecorder) WriteHeader(code int)        { r.code = code }
func (r *responseRecorder) Write(b []byte) (int, error) { return r.buf.Write(b) }

func (r *responseRecorder) reset() {
	if r.buf != nil {
		r.buf.Reset()
		respBufPool.Put(r.buf)
		r.buf = nil
	}
}

// ─── writeJSONError ──────────────────────────────────────────────────────────

func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
