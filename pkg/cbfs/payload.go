// Copyright 2018-2021 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cbfs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

func init() {
	if err := RegisterFileReader(&SegReader{Type: TypeSELF, Name: "Payload", New: NewPayloadRecord}); err != nil {
		log.Fatal(err)
	}
}

func NewPayloadRecord(f *File) (ReadWriter, error) {
	p := &PayloadRecord{File: *f}
	return p, nil
}

func (r *PayloadRecord) Read(in io.ReadSeeker) error {
	for {
		var h PayloadHeader
		if err := Read(in, &h); err != nil {
			Debug("PayloadHeader read: %v", err)
			return err
		}
		Debug("Got PayloadHeader %s", h.String())
		r.Segs = append(r.Segs, h)
		if h.Type == SegEntry {
			break
		}
	}
	// Seek to offset (after the header); the remainder is the actual payload.
	offset, err := in.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("finding location in stream: %w", err)
	}
	bodySize := int64(r.Size) - offset
	Debug("Payload size: %v, body size: %v, offset: %v", r.Size, bodySize, offset)
	if bodySize < 0 {
		// This should not happen. Tolerate a potential error.
		return nil
	}
	// This _may_ happen. E.g. with the test payload here. Silently ignore.
	if bodySize == 0 {
		Debug("Payload empty, nothing to read")
		return nil
	}
	r.FData = make([]byte, bodySize)
	n, err := in.Read(r.FData)
	if err != nil {
		return err
	}
	Debug("Payload read %d bytes", n)
	return nil
}

// struct for PayloadRecord marshalling
type mPayloadRecord struct {
	Name        string
	Start       uint32
	Size        uint32
	Type        string
	Segments    []PayloadHeader
	Compression string
}

func (r *PayloadRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(mPayloadRecord{
		Name:        r.Name,
		Start:       r.RecordStart,
		Size:        r.FileHeader.Size,
		Type:        r.FileHeader.Type.String(),
		Segments:    r.Segs,
		Compression: r.File.Compression().String(),
	})
}

func (r *PayloadRecord) String() string {
	s := recString(r.File.Name, r.RecordStart, r.Type.String(), r.Size, "none")
	for i, seg := range r.Segs {
		s += "\n"
		s += recString(fmt.Sprintf(" Seg #%d", i), seg.Offset, seg.Type.String(), seg.Size, seg.Compression.String())
	}
	return s
}

// struct for PayloadHeader marshalling
type mPayloadHeader struct {
	Type        string
	Compression string
	Offset      uint32
	LoadAddress uint64
	Size        uint32
	MemSize     uint32
}

func (h *PayloadHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(mPayloadHeader{
		Type:        h.Type.String(),
		Compression: h.Compression.String(),
		Offset:      h.Offset,
		LoadAddress: h.LoadAddress,
		Size:        h.Size,
		MemSize:     h.MemSize,
	})
}

func (h *PayloadHeader) String() string {
	return fmt.Sprintf("Type %#x Compression %#x Offset %#x LoadAddress %#x Size %#x MemSize %#x",
		h.Type,
		h.Compression,
		h.Offset,
		h.LoadAddress,
		h.Size,
		h.MemSize)
}

func (r *PayloadRecord) Write(w io.Writer) error {
	if err := Write(w, r.Segs); err != nil {
		return err
	}
	return Write(w, r.FData)
}

func (r *PayloadRecord) GetFile() *File {
	return &r.File
}
