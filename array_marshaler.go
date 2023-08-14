package zap_human

import (
	"encoding/hex"
	"go.uber.org/zap/zapcore"
	"strings"
	"time"
)

func (h *HumanEncoder) AppendBool(b bool) {
	h.writeStringf("%T", b)
}

func (h *HumanEncoder) AppendByteString(bytes []byte) {
	lines := strings.Split(hex.Dump(bytes), "\n")
	for i := range lines {
		lines[i] = h.ident() + lines[i]
	}
	h.writeString(strings.Join(lines, "\n") + "\n")
}

func (h *HumanEncoder) AppendComplex128(c complex128) { h.appendComplex(c, 64) }

func (h *HumanEncoder) AppendComplex64(c complex64) { h.appendComplex(complex128(c), 32) }

func (h *HumanEncoder) AppendFloat64(f float64) { h.appendFloat(f, 64) }

func (h *HumanEncoder) AppendFloat32(f float32) { h.appendFloat(float64(f), 32) }

func (h *HumanEncoder) AppendInt(i int) { h.AppendInt64(int64(i)) }

func (h *HumanEncoder) AppendInt64(i int64) { h.writeStringf("%d", i) }

func (h *HumanEncoder) AppendInt32(i int32) { h.AppendInt64(int64(i)) }

func (h *HumanEncoder) AppendInt16(i int16) { h.AppendInt64(int64(i)) }

func (h *HumanEncoder) AppendInt8(i int8) { h.AppendInt64(int64(i)) }

func (h *HumanEncoder) AppendString(s string) { h.writeString(s) }

func (h *HumanEncoder) AppendUint(u uint) { h.AppendUint64(uint64(u)) }

func (h *HumanEncoder) AppendUint64(u uint64) { h.writeStringf("%d", u) }

func (h *HumanEncoder) AppendUint32(u uint32) { h.AppendUint64(uint64(u)) }

func (h *HumanEncoder) AppendUint16(u uint16) { h.AppendUint64(uint64(u)) }

func (h *HumanEncoder) AppendUint8(u uint8) { h.AppendUint64(uint64(u)) }

func (h *HumanEncoder) AppendUintptr(u uintptr) { h.AppendUint64(uint64(u)) }

func (h *HumanEncoder) AppendDuration(duration time.Duration) { h.writeString(duration.String()) }

func (h *HumanEncoder) AppendTime(t time.Time) {
	h.writeStringf("%s", t.Format(time.RFC3339Nano))
}

func (h *HumanEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	h.writeString("[ ")
	err := marshaler.MarshalLogArray(h)
	h.writeString(" ]\n")
	return err
}

func (h *HumanEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	old := h.openNamespaces
	h.openNamespaces = 0
	h.buf.AppendByte('{')
	err := marshaler.MarshalLogObject(h)
	h.writeString("}\n")
	h.openNamespaces = old
	return err
}

func (h *HumanEncoder) AppendReflected(value interface{}) error {
	valueBytes, err := h.encodeReflected(value)
	if err != nil {
		return err
	}
	_, err = h.buf.Write(valueBytes)
	return err
}
