package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/eyecuelab/kit/maputil"
	"github.com/eyecuelab/kit/set"
	"github.com/google/jsonapi"
	"github.com/labstack/echo"
)

var reNotJsonApi = regexp.MustCompile("not a jsonapi|EOF")

func notJsonApi(err error) bool {
	return reNotJsonApi.MatchString(err.Error())
}

type (
	ApiContext interface {
		echo.Context

		Payload() *jsonapi.OnePayload
		Attrs() map[string]interface{}
		AttrKeys() []string
		BindAndValidate(interface{}) error
		BindIdParam(*int, ...string) error
		JsonApi(interface{}, int) error
		JsonApiOK(interface{}) error
		ApiError(string, ...int) *echo.HTTPError
		RestrictedParam(string, ...string) (string, error)
		QueryParamTrue(string) (bool, bool)
		RequiredQueryParams(...string) (map[string]string, error)
		OptionalQueryParams(...string) map[string]string
		QParams(...string) (map[string]string, error)
	}

	apiContext struct {
		echo.Context

		payload *jsonapi.OnePayload
	}
)

func (c *apiContext) Payload() *jsonapi.OnePayload {
	return c.payload
}

func (c *apiContext) Attrs() map[string]interface{} {
	return c.payload.Data.Attributes
}

func (c *apiContext) AttrKeys() []string {
	return maputil.Keys(c.Attrs())
}

func (c *apiContext) Bind(i interface{}) error {
	ctype := c.Request().Header.Get(echo.HeaderContentType)

	if isJSONAPI(ctype) {
		return jsonAPIBind(c, i)
	}
	return c.defaultBind(i)
}

func (c *apiContext) defaultBind(i interface{}) error {
	db := new(echo.DefaultBinder)
	return db.Bind(i, c)
}

func isJSONAPI(s string) bool {
	const MIMEJsonAPI = "application/vnd.api+json"
	return strings.HasPrefix(s, MIMEJsonAPI)
}

func (c *apiContext) BindAndValidate(i interface{}) error {
	if err := c.Bind(i); err != nil {
		return err
	}
	if err := c.Validate(i); err != nil {
		return err
	}
	return nil
}

func (c *apiContext) JsonApi(i interface{}, status int) error {
	c.Response().Header().Set(echo.HeaderContentType, jsonapi.MediaType)
	c.Response().WriteHeader(status)

	return jsonapi.MarshalPayload(c.Response().Writer, i)
}

func (c *apiContext) JsonApiOK(i interface{}) error {
	return c.JsonApi(i, http.StatusOK)
}

func (c *apiContext) BindIdParam(idValue *int, named ...string) (err error) {
	paramName := "id"
	if len(named) > 0 {
		paramName = named[0]
	}
	*idValue, err = strconv.Atoi(c.Param(paramName))
	return err
}

func (c *apiContext) QueryParamTrue(name string) (val, ok bool) {
	switch strings.ToLower(c.QueryParam(name)) {
	case "true", "1":
		return true, true
	case "false", "0":
		return false, true
	default:
		return false, false
	}
}

func jsonAPIBind(c *apiContext, i interface{}) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(c.Request().Body, buf)

	if err := jsonapi.UnmarshalPayload(tee, i); err != nil {
		if notJsonApi(err) {
			return c.ApiError("Request Body is not valid JsonAPI")
		}
		return err
	}

	c.payload = new(jsonapi.OnePayload)
	return json.Unmarshal(buf.Bytes(), c.payload)
}

func (c *apiContext) ApiError(msg string, codes ...int) *echo.HTTPError {
	if len(codes) > 0 {
		return echo.NewHTTPError(codes[0], msg)
	}
	// TODO: return jsonapi error instead
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

func (c *apiContext) RestrictedParam(paramName string, allowedValues ...string) (string, error) {
	return restrictedValue(c.Param(paramName), allowedValues, "Param value %v not allowed")
}

func (c *apiContext) RestrictedQueryParam(paramName string, allowedValues ...string) (string, error) {
	return restrictedValue(c.QueryParam(paramName), allowedValues, "Query param value %v not allowed")
}

func (c *apiContext) RequiredQueryParams(required ...string) (map[string]string, error) {
	missing := make([]string, 0, len(required))
	params := make(map[string]string)

	for _, key := range required {
		val := c.QueryParam(key)
		if val == "" {
			missing = append(missing, key)
			continue
		}
		params[key] = val
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required params: %v", missing)
	}

	return params, nil
}

func (c *apiContext) QParams(required ...string) (map[string]string, error) {

	missing := make([]string, 0, len(required))
	params := make(map[string]string)

	for k := range c.QueryParams() {
		params[k] = c.QueryParam(k)
	}

	for _, k := range required {
		if _, ok := params[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required params: %v", missing)
	}

	return params, nil
}

func stringSet(m url.Values) set.String {
	set := make(set.String)
	for k := range m {
		set.Add(k)
	}
	return set
}

func (c *apiContext) OptionalQueryParams(optional ...string) map[string]string {
	params := make(map[string]string)
	for _, key := range optional {
		val := c.QueryParam(key)
		params[key] = val
	}
	return params
}

func ApiContextMiddleWare() func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ac := &apiContext{c, nil}
			return next(ac)
		}
	}
}

func restrictedValue(value string, allowed []string, errorText string) (string, error) {
	if contains(allowed, value) {
		return value, nil
	}
	return "", fmt.Errorf(errorText, value)
}

func contains(set []string, s string) bool {
	for _, v := range set {
		if s == v {
			return true
		}
	}
	return false
}
