package custom_nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/shared"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type (
	Handler interface{}
)

var (
	ContextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	ErrorType   = reflect.TypeOf((*error)(nil)).Elem()
)

type Router struct {
	http.Handler
	chi chi.Router
}

func NewRouter(chi chi.Router) *Router {
	return &Router{
		chi: chi,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.chi.ServeHTTP(w, r)
}

func decode(data []byte, vPtr any) error {
	switch arg := vPtr.(type) {
	case *string:
		str := string(data)
		if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
			*arg = str[1 : len(str)-1]
		} else {
			*arg = str
		}
	case *[]byte:
		*arg = data
	default:
		if err := json.Unmarshal(data, arg); err != nil {
			return err
		}
	}
	return nil
}

func (router *Router) doRequest(r *http.Request, h Handler, ctx context.Context) (interface{}, error) {
	// reflect h to a func in go --> call it
	if h == nil {
		return nil, errors.New("NATS: handler is required")
	}
	handlerType := reflect.TypeOf(h)
	if handlerType.Kind() != reflect.Func {
		return nil, errors.New("handler must be func with format func(ctx *Context, req interface{}) (interface{}, err) or func(ctx *Context, req interface{}) error")
	}

	numIn := handlerType.NumIn()
	numOut := handlerType.NumOut()

	if numIn == 0 || numIn > 2 {
		return nil, errors.New("Handler requires one or two parameters")
	}

	firstInParameter := handlerType.In(0)
	if firstInParameter != ContextType {
		return nil, errors.New("handler muste have first parameter is a instance of context.Context")
	}

	if numOut == 0 || numOut > 2 {
		return nil, errors.New("Handler must have one or two output value")
	}

	if handlerType.Out(numOut-1) != ErrorType {
		return nil, errors.New("Handler must return a error")
	}

	reqType := handlerType.In(numIn - 1)
	if reqType == nil {
		return nil, errors.New("handler must be have on request parameter")
	}
	// responseType := handlerType.Out(0)

	handlerValue := reflect.ValueOf(h)

	input := []reflect.Value{reflect.ValueOf(ctx)}
	bodyReader := r.Body
	defer bodyReader.Close()

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	if numIn == 2 {
		if len(body) == 0 {
			return nil, errors.New("body is empty")
		}

		var req reflect.Value
		if reqType.Kind() != reflect.Pointer {
			req = reflect.New(reqType)
		} else {
			req = reflect.New(reqType.Elem())
		}

		if err := decode(body, req.Interface()); err != nil {
			return nil, err
		}
		input = append(input, req)
	}

	res := handlerValue.Call(input)

	var errorReturn error
	if numOut == 2 {
		ret := res[0].Interface()
		if v := res[1].Interface(); v != nil {
			errorReturn = v.(error)
		}

		return ret, errorReturn
	} else { // numout = 1
		if v := res[0].Interface(); v != nil {
			errorReturn := v.(error)
			return nil, errorReturn
		}
	}
	return nil, nil
}

func (router *Router) handlerRequest(r *http.Request, h Handler, ctx context.Context) (*Response, error) {
	returnValue, errFromAPI := router.doRequest(r, h, ctx)
	responseBuilder := NewResponseBuilder(http.StatusOK).BuildHeader(r.Header)
	if errFromAPI != nil {
		return nil, errFromAPI
	}

	switch returnType := returnValue.(type) {
	case string:
		return responseBuilder.BuildBody([]byte(returnValue.(string))).Build(), nil
	case int, int16, int32, int64, int8, float32, float64, bool:
		return responseBuilder.BuildBody([]byte(fmt.Sprintf("%v", returnValue))).Build(), nil
	case *Response:
		return returnValue.(*Response), nil
	default:
		_ = returnType
		resJson, err := json.Marshal(returnValue)
		if err != nil {
			return nil, err
		}

		return responseBuilder.BuildBody(resJson).Build(), nil
	}
}

func (router *Router) RegisterRoute(method, path string, h Handler) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), shared.HTTPRequest_ContextKey, r)
		ctx = context.WithValue(ctx, shared.HTTPResponse_ContextKey, w)
		// additional info to context
		// We will build context here
		// 1: Get from header
		userId := r.Header.Get("X-User-Id")
		if userId != "" {
			ctx = context.WithValue(ctx, shared.UserId_ContextKey, userId)
		}

		res, err := router.handlerRequest(r, h, ctx)
		if err != nil {
			w.Header().Set("Content-type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)

			respJson := map[string]string{
				"error": err.Error(),
			}
			jsonByte, err := json.Marshal(respJson)
			if err != nil {
				log.Default().Println("fail to marshal response ", err)
			}

			if _, err := w.Write(jsonByte); err != nil {
				log.Default().Println("fail to send err response json")
			}
			return
		}

		if res == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for key, headers := range res.Headers {
			for _, item := range headers {
				w.Header().Add(key, item)
			}
		}
		w.WriteHeader(int(res.StatusCode))
		bodyRes := res.Body
		if len(bodyRes) > 0 {
			_, err := w.Write(bodyRes)
			if err != nil {
				log.Default().Println("fail to send err response json")
			}
		}
	}
	router.chi.Method(method, path, otelhttp.NewHandler(
		http.HandlerFunc(handler),
		path,
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	))
}
