package zen

// #include "zen_engine.h"
import "C"
import (
	"encoding/json"
	"errors"
	"github.com/tidwall/gjson"
)

type CustomNodeHandler func(request NodeRequest) (NodeResponse, error)

type CustomNode struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Kind   string          `json:"kind"`
	Config json.RawMessage `json:"config"`
}

type NodeRequest struct {
	Node  CustomNode      `json:"node"`
	Input json.RawMessage `json:"input"`
}

type NodeResponse struct {
	Output    any `json:"output"`
	TraceData any `json:"traceData"`
}

func wrapCustomNodeHandler(customNodeHandler CustomNodeHandler) func(cRequest *C.char) C.ZenCustomNodeResult {
	return func(cRequest *C.char) C.ZenCustomNodeResult {
		strRequest := C.GoString(cRequest)

		var request NodeRequest
		if err := json.Unmarshal([]byte(strRequest), &request); err != nil {
			return C.ZenCustomNodeResult{
				content: nil,
				error:   C.CString(err.Error()),
			}
		}

		response, err := customNodeHandler(request)
		if err != nil {
			return C.ZenCustomNodeResult{
				content: nil,
				error:   C.CString(err.Error()),
			}
		}

		cResponse, err := json.Marshal(response)
		if err != nil {
			return C.ZenCustomNodeResult{
				content: nil,
				error:   C.CString(err.Error()),
			}
		}

		return C.ZenCustomNodeResult{
			content: C.CString(string(cResponse)),
			error:   nil,
		}
	}
}

func GetNodeFieldRaw[T any](request NodeRequest, path string) (T, error) {
	result := gjson.GetBytes(request.Node.Config, path)
	if !result.Exists() {
		return *new(T), errors.New("path does not exist")
	}

	var r T
	if err := json.Unmarshal([]byte(result.Raw), &r); err != nil {
		return *new(T), err
	}

	return r, nil
}

func GetNodeField[T any](request NodeRequest, path string) (T, error) {
	result := gjson.GetBytes(request.Node.Config, path)
	if !result.Exists() {
		return *new(T), errors.New("path does not exist")
	}

	if result.Type != gjson.String {
		var r T
		if err := json.Unmarshal([]byte(result.Raw), &r); err != nil {
			return *new(T), err
		}

		return r, nil
	}

	return RenderTemplate[T](result.Str, request.Input)
}
