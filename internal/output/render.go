package output

import (
	"encoding/json"
	"fmt"
	"io"
)

func Render(w io.Writer, format string, text string, data any) error {
	if format == "json" {
		payload, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(payload))
		return err
	}

	if text == "" {
		return nil
	}
	_, err := fmt.Fprintln(w, text)
	return err
}

func RenderError(w io.Writer, err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintln(w, err.Error())
}
