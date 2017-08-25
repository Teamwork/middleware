package log

type errorWithFields struct {
	err    error
	fields Fields
}

type errorHasFields interface {
	Fields() Fields
}

func (e *errorWithFields) Cause() error {
	return e.err
}

func (e *errorWithFields) Fields() Fields {
	return e.fields
}

// ExtractFields extracts all fields from the error, and any wrapped errors.
// Newer/higher fields take precident over earlier ones, in case of collision.
func ExtractFields(err error) Fields {
	haveFields := make([]errorHasFields, 0)
	for {
		if fieldErr, ok := err.(errorHasFields); ok {
			haveFields = append(haveFields, fieldErr)
		}
		if cause, ok := err.(causer); ok {
			err = cause.Cause()
		} else {
			break
		}
	}
	if len(haveFields) == 0 {
		return nil
	}
	fields := make(map[string]interface{})
	for i := len(haveFields) - 1; i >= 0; i-- {
		for k, v := range haveFields[i].Fields() {
			fields[k] = v
		}
	}
	return fields
}

func (e *errorWithFields) Error() string {
	return e.err.Error()
}

// ErrorWithFields bundles an existing error with fields, which can be extracted
// and used by the logger.
func ErrorWithFields(err error, fields Fields) error {
	if err == nil {
		return nil
	}
	return &errorWithFields{
		err:    err,
		fields: fields,
	}
}

// ErrorWithFields bundles an existing error with fields, which can be extracted
// and used by the logger. Any fields already set in the Entry are included, and
// may be overwritten by fields passed into this function.
func (e *Entry) ErrorWithFields(err error, fields Fields) error {
	data := make(Fields)
	for k, v := range e.e.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return ErrorWithFields(err, data)
}
