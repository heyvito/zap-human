package zap_human

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"io"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	_pool         = buffer.NewPool()
	GetBufferPool = _pool.Get
)

var loggerPool = NewPool(func() *HumanEncoder {
	return &HumanEncoder{}
})

func NewHumanEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return newHumanEncoder(cfg), nil
}

func defaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	// For consistency with our custom JSON encoder.
	enc.SetEscapeHTML(false)
	return enc
}

func newHumanEncoder(cfg zapcore.EncoderConfig) *HumanEncoder {
	if cfg.SkipLineEnding {
		cfg.LineEnding = ""
	} else if cfg.LineEnding == "" {
		cfg.LineEnding = zapcore.DefaultLineEnding
	}

	// If no EncoderConfig.NewReflectedEncoder is provided by the user, then use default
	if cfg.NewReflectedEncoder == nil {
		cfg.NewReflectedEncoder = defaultReflectedEncoder
	}

	return &HumanEncoder{
		EncoderConfig: &cfg,
		buf:           GetBufferPool(),
	}
}

type HumanEncoder struct {
	*zapcore.EncoderConfig
	buf            *buffer.Buffer
	openNamespaces int
	reflectBuf     *buffer.Buffer
	reflectEnc     zapcore.ReflectedEncoder
}

func (h *HumanEncoder) ident() string {
	return "                                      \t" +
		strings.Repeat("  ", h.openNamespaces)
}

func (h *HumanEncoder) writeString(str string) {
	_, _ = h.buf.WriteString(str)
}

func (h *HumanEncoder) writeStringf(f string, a ...any) {
	_, _ = h.buf.WriteString(fmt.Sprintf(f, a...))
}

func (h *HumanEncoder) addKey(k string) {
	h.writeString(h.ident() + k + ": ")
}

func (h *HumanEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	h.addKey(key)
	h.writeString("[ ")
	err := marshaler.MarshalLogArray(h)
	h.writeString(" ]\n")
	return err
}

func (h *HumanEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	h.addKey(key)
	old := h.openNamespaces
	h.openNamespaces = 0
	h.buf.AppendByte('{')
	err := marshaler.MarshalLogObject(h)
	h.writeString("}\n")
	h.openNamespaces = old
	return err
}

func (h *HumanEncoder) AddBinary(key string, value []byte) {
	h.addKey(key)
	h.buf.AppendByte('\n')
	h.openNamespaces++
	lines := strings.Split(hex.Dump(value), "\n")
	for i := range lines {
		lines[i] = h.ident() + lines[i]
	}
	h.writeString(strings.Join(lines, "\n") + "\n")
	h.openNamespaces--
}

func (h *HumanEncoder) AddByteString(key string, value []byte) {
	h.AddBinary(key, value)
}

func (h *HumanEncoder) AddBool(key string, value bool) {
	h.addKey(key)
	h.AppendBool(value)
	_ = h.buf.WriteByte('\n')
}

func (h *HumanEncoder) AddComplex128(key string, value complex128) {
	h.addKey(key)
	h.appendComplex(value, 64)
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) AddComplex64(key string, value complex64) {
	h.addKey(key)
	h.appendComplex(complex128(value), 32)
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) appendComplex(val complex128, precision int) {
	r, i := real(val), imag(val)
	h.buf.AppendFloat(r, precision)
	if i >= 0 {
		h.buf.AppendByte('+')
	}
	h.buf.AppendFloat(i, precision)
	h.buf.AppendByte('i')
}

func (h *HumanEncoder) AddDuration(key string, value time.Duration) {
	h.addKey(key)
	h.writeString(value.String())
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) AddFloat64(key string, value float64) {
	h.addKey(key)
	h.appendFloat(value, 64)
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) AddFloat32(key string, value float32) {
	h.addKey(key)
	h.appendFloat(float64(value), 32)
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		h.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		h.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		h.buf.AppendString(`"-Inf"`)
	default:
		h.buf.AppendFloat(val, bitSize)
	}
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) AddInt(key string, value int) { h.AddInt64(key, int64(value)) }

func (h *HumanEncoder) AddInt64(key string, value int64) {
	h.addKey(key)
	h.writeStringf("%d\n", value)
}

func (h *HumanEncoder) AddInt32(key string, value int32) { h.AddInt64(key, int64(value)) }

func (h *HumanEncoder) AddInt16(key string, value int16) { h.AddInt64(key, int64(value)) }

func (h *HumanEncoder) AddInt8(key string, value int8) { h.AddInt64(key, int64(value)) }

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func trimRightSpace(s string) string {
	start := 0
	stop := len(s)
	for ; stop > start; stop-- {
		c := s[stop-1]
		if c >= utf8.RuneSelf {
			return strings.TrimRightFunc(s[start:stop], unicode.IsSpace)
		}
		if asciiSpace[c] == 0 {
			break
		}
	}

	return s[start:stop]
}

func (h *HumanEncoder) AddString(key, value string) {
	h.addKey(key)
	if strings.ContainsRune(value, '\n') {
		h.openNamespaces++
		lines := strings.Split(value, "\n")
		for i := range lines {
			lines[i] = h.ident() + trimRightSpace(lines[i])
		}
		h.openNamespaces--
		h.writeString("\n" + strings.Join(lines, "\n") + "\n")
	} else {
		h.writeStringf("%s\n", value)
	}
}

func (h *HumanEncoder) AddTime(key string, value time.Time) {
	h.addKey(key)
	h.writeStringf("%s\n", value.Format(time.RFC3339Nano))
}

func (h *HumanEncoder) AddUint(key string, value uint) { h.AddUint64(key, uint64(value)) }

func (h *HumanEncoder) AddUint64(key string, value uint64) {
	h.addKey(key)
	h.writeStringf("%d\n", value)
}

func (h *HumanEncoder) AddUint32(key string, value uint32) { h.AddUint64(key, uint64(value)) }

func (h *HumanEncoder) AddUint16(key string, value uint16) { h.AddUint64(key, uint64(value)) }

func (h *HumanEncoder) AddUint8(key string, value uint8) { h.AddUint64(key, uint64(value)) }

func (h *HumanEncoder) AddUintptr(key string, value uintptr) { h.AddUint64(key, uint64(value)) }

func (h *HumanEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return []byte("nil"), nil
	}
	h.resetReflectBuf()
	if err := h.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}
	h.reflectBuf.TrimNewline()
	return h.reflectBuf.Bytes(), nil
}

func (h *HumanEncoder) AddReflected(key string, value interface{}) error {
	valueBytes, err := h.encodeReflected(value)
	if err != nil {
		return err
	}
	h.addKey(key)
	_, err = h.buf.Write(valueBytes)
	h.writeString("\n")
	return err
}

func (h *HumanEncoder) OpenNamespace(key string) {
	h.openNamespaces++
	h.addKey(key)
	h.buf.AppendByte('\n')
}

func (h *HumanEncoder) Clone() zapcore.Encoder {
	clone := *h
	clone.buf = &buffer.Buffer{}
	_, _ = clone.buf.Write(h.buf.Bytes())
	return &clone
}

func (h *HumanEncoder) clone() *HumanEncoder {
	clone := loggerPool.Get()
	clone.EncoderConfig = h.EncoderConfig
	clone.openNamespaces = h.openNamespaces
	clone.buf = GetBufferPool()
	return clone
}

func (h *HumanEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := h.clone()

	if final.EncodeLevel != nil {
		final.EncodeLevel(ent.Level, final)
		final.writeString(" ")
	}

	final.AppendTime(ent.Time)
	final.writeString(" ")

	final.AppendString(ent.LoggerName)
	final.writeString("\t")

	if ent.Caller.Defined {
		final.EncodeCaller(ent.Caller, final)
		final.writeString("\t")
		if final.FunctionKey != "" {
			final.AppendString(ent.Caller.Function)
			final.writeString("\t")
		}
	}

	final.AppendString(ent.Message)
	final.writeString("\n")

	if h.buf.Len() > 0 {
		_, _ = final.buf.Write(h.buf.Bytes())
	}
	addFields(final, fields)
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.AddString(final.StacktraceKey, ent.Stack)
	}
	final.writeString("\n")

	ret := final.buf
	putEncoder(final)
	return ret, nil
}

func putEncoder(enc *HumanEncoder) {
	if enc.reflectBuf != nil {
		enc.reflectBuf.Free()
	}
	enc.EncoderConfig = nil
	enc.buf = nil
	enc.openNamespaces = 0
	enc.reflectBuf = nil
	enc.reflectEnc = nil
	loggerPool.Put(enc)
}

func (h *HumanEncoder) resetReflectBuf() {
	if h.reflectBuf == nil {
		h.reflectBuf = GetBufferPool()
		h.reflectEnc = h.NewReflectedEncoder(h.reflectBuf)
	} else {
		h.reflectBuf.Reset()
	}
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
