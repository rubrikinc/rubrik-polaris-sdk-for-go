package cdm

import "testing"

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		res  []byte
		code int
		msg  string
	}{{
		name: "EmptyResponse",
		res:  []byte(``),
		code: 404,
		msg:  "Not Found (404)",
	}, {
		name: "TextResponse",
		res:  []byte(`error message`),
		code: 500,
		msg:  "Internal Server Error (500): error message",
	}, {
		name: "EmptyJSON",
		res:  []byte(`{}`),
		code: 503,
		msg:  "Service Unavailable (503): {}",
	}, {
		name: "JSONResponse",
		res:  []byte(`{"name":"value"}`),
		code: 403,
		msg:  `Forbidden (403): {"name":"value"}`,
	}, {
		name: "PartialJSONErrorResponse",
		res:  []byte(`{"errorType":"errorType"}`),
		code: 500,
		msg:  `Internal Server Error (500): {"errorType":"errorType"}`,
	}, {
		name: "FullJSONErrorResponse",
		res:  []byte(`{"errorType":"errorType","message":"message"}`),
		code: 404,
		msg:  "Not Found (404): errorType: message",
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if msg := errorMessage(testCase.res, testCase.code); msg != testCase.msg {
				t.Fatalf("%s != %s", msg, testCase.msg)
			}
		})
	}
}
