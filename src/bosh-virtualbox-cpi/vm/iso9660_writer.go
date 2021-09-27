package vm

// more or less vendored from github.com/johto/iso9660wrap/blob/master/iso9660_writer.go

import (
	"encoding/binary"
	"io"
	"math"
	"strings"
	"time"
	"fmt"
)

const SectorSize uint32 = 2048

type SectorWriter struct {
	w io.Writer
	p uint32
}

func (w *SectorWriter) Write(p []byte) (uint32, error) {
	if len(p) >= math.MaxUint32 {
		return 0, fmt.Errorf("attempted write of length %d is out of sector bounds", len(p))
	}
	l := uint32(len(p))
	if l > w.RemainingSpace() {
		return 0, fmt.Errorf("attempted write of length %d at offset %d is out of sector bounds", w.p, len(p))
	}
	w.p += l
	_, err := w.w.Write(p)
		if err != nil {
			return 0, err
		}
	return l, nil
}

func (w *SectorWriter) WriteUnspecifiedDateTime() (uint32, error) {
	b := make([]byte, 17)
	for i := 0; i < 16; i++ {
		b[i] = '0'
	}
	b[16] = 0
	return w.Write(b)
}

func (w *SectorWriter) WriteDateTime(t time.Time) (uint32, error) {
	f := t.UTC().Format("20060102150405")
	f += "00" // 1/100
	f += "\x00" // UTC offset
	if len(f) != 17 {
		return 0, fmt.Errorf("date and time field %q is of unexpected length %d", f, len(f))
	}
	return w.WriteString(f)
}

func (w *SectorWriter) WriteString(str string) (uint32, error) {
	return w.Write([]byte(str))
}

func (w *SectorWriter) WritePaddedString(str string, length uint32) (uint32, error) {
	l, err := w.WriteString(str)
	if err != nil {
		return 0, err
	}
	if l > 32 {
		return 0, fmt.Errorf("padded string %q exceeds length %d", str, length)
	} else if l < 32 {
		w.WriteString(strings.Repeat(" ", int(32 - l)))
	}
	return 32, nil
}

func (w *SectorWriter) WriteByte(b byte) (uint32, error) {
	return w.Write([]byte{b})
}

func (w *SectorWriter) WriteWord(bo binary.ByteOrder, word uint16) (uint32, error) {
	b := make([]byte, 2)
	bo.PutUint16(b, word)
	return w.Write(b)
}

func (w *SectorWriter) WriteBothEndianWord(word uint16) (uint32, error) {
	w.WriteWord(binary.LittleEndian, word)
	w.WriteWord(binary.BigEndian, word)
	return 4, nil
}

func (w *SectorWriter) WriteDWord(bo binary.ByteOrder, dword uint32) (uint32, error) {
	b := make([]byte, 4)
	bo.PutUint32(b, dword)
	return w.Write(b)
}

func (w *SectorWriter) WriteLittleEndianDWord(dword uint32) (uint32, error) {
	return w.WriteDWord(binary.LittleEndian, dword)
}

func (w *SectorWriter) WriteBigEndianDWord(dword uint32) (uint32, error) {
	return w.WriteDWord(binary.BigEndian, dword)
}

func (w *SectorWriter) WriteBothEndianDWord(dword uint32) uint32 {
	w.WriteLittleEndianDWord(dword)
	w.WriteBigEndianDWord(dword)
	return 8
}

func (w *SectorWriter) WriteZeros(c int) (uint32, error) {
	return w.Write(make([]byte, c))
}

func (w *SectorWriter) PadWithZeros() (uint32, error) {
	return w.Write(make([]byte, w.RemainingSpace()))
}

func (w *SectorWriter) RemainingSpace() uint32 {
	return SectorSize - w.p
}

func (w *SectorWriter) Reset() {
	w.p = 0
}


type ISO9660Writer struct {
	sw *SectorWriter
	sectorNum uint32
}

func (w *ISO9660Writer) CurrentSector() uint32 {
	return uint32(w.sectorNum)
}

func (w *ISO9660Writer) NextSector() (*SectorWriter, error) {
	if w.sw.RemainingSpace() == SectorSize {
		return nil, fmt.Errorf("internal error: tried to leave sector %d empty", w.sectorNum)
	}
	w.sw.PadWithZeros()
	w.sw.Reset()
	w.sectorNum++
	return w.sw, nil
}

func (w *ISO9660Writer) Finish() {
	if w.sw.RemainingSpace() != SectorSize {
		w.sw.PadWithZeros()
	}
	w.sw = nil
}

func NewISO9660Writer(w io.Writer) *ISO9660Writer {
	// start at the end of the last reserved sector
	return &ISO9660Writer{&SectorWriter{w, SectorSize}, 16 - 1}
}
