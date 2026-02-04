package mongo

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

var (
	tUUID    = reflect.TypeFor[uuid.UUID]()
	tPtrUUID = reflect.TypeFor[*uuid.UUID]()
)

// uuidCodec handles encoding/decoding of uuid.UUID to/from BSON Binary subtype 4
type uuidCodec struct{}

func (c *uuidCodec) EncodeValue(_ bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{
			Name:     "uuidCodec",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}

	uid := val.Interface().(uuid.UUID)
	// Write as Binary subtype 4 (UUID)
	return vw.WriteBinaryWithSubtype(uid[:], bson.TypeBinaryUUID)
}

func (c *uuidCodec) DecodeValue(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{
			Name:     "uuidCodec",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}

	var data []byte
	var subtype byte
	var err error

	switch vr.Type() {
	case bson.TypeBinary:
		data, subtype, err = vr.ReadBinary()
		if err != nil {
			return err
		}
		// Accept both subtype 0 (legacy) and subtype 4 (UUID standard)
		if subtype != 0x00 && subtype != 0x04 {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bson.TypeNull:
		return vr.ReadNull()
	case bson.TypeUndefined:
		return vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a uuid.UUID", vr.Type())
	}

	if len(data) != 16 {
		return fmt.Errorf("invalid UUID length: %d", len(data))
	}

	uid, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uid))
	return nil
}

// ptrUUIDCodec handles encoding/decoding of *uuid.UUID
type ptrUUIDCodec struct {
	codec *uuidCodec
}

func (c *ptrUUIDCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tPtrUUID {
		return bsoncodec.ValueEncoderError{
			Name:     "ptrUUIDCodec",
			Types:    []reflect.Type{tPtrUUID},
			Received: val,
		}
	}

	if val.IsNil() {
		return vw.WriteNull()
	}

	return c.codec.EncodeValue(ec, vw, val.Elem())
}

func (c *ptrUUIDCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tPtrUUID {
		return bsoncodec.ValueDecoderError{
			Name:     "ptrUUIDCodec",
			Types:    []reflect.Type{tPtrUUID},
			Received: val,
		}
	}

	if vr.Type() == bson.TypeNull {
		val.Set(reflect.Zero(tPtrUUID))
		return vr.ReadNull()
	}

	if vr.Type() == bson.TypeUndefined {
		val.Set(reflect.Zero(tPtrUUID))
		return vr.ReadUndefined()
	}

	if val.IsNil() {
		val.Set(reflect.New(tUUID))
	}

	return c.codec.DecodeValue(dc, vr, val.Elem())
}

// NewRegistryWithUUID creates a new BSON registry with UUID codec registered
func NewRegistryWithUUID() *bsoncodec.Registry {
	rb := bson.NewRegistry()
	codec := &uuidCodec{}
	ptrCodec := &ptrUUIDCodec{codec: codec}

	rb.RegisterTypeEncoder(tUUID, codec)
	rb.RegisterTypeDecoder(tUUID, codec)
	rb.RegisterTypeEncoder(tPtrUUID, ptrCodec)
	rb.RegisterTypeDecoder(tPtrUUID, ptrCodec)

	return rb
}
