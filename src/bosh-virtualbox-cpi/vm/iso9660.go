package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

// More or less vendored from https://github.com/johto/iso9660wrap/blob/master/iso9660wrap.go
type ISO9660 struct {
	FileName string
	Contents []byte
}

const (
	volumeDescriptorSetMagic              = "\x43\x44\x30\x30\x31\x01"
	primaryVolumeSectorNum         uint32 = 16
	numVolumeSectors               uint32 = 2 // primary + terminator
	littleEndianPathTableSectorNum uint32 = primaryVolumeSectorNum + numVolumeSectors
	bigEndianPathTableSectorNum    uint32 = littleEndianPathTableSectorNum + 1
	numPathTableSectors                   = 2 // no secondaries
	rootDirectorySectorNum         uint32 = primaryVolumeSectorNum + numVolumeSectors + numPathTableSectors
)

func (i ISO9660) inputLen() uint32 {
	return uint32(len(i.Contents))
}

func (i ISO9660) Bytes() ([]byte, error) {
	i.FileName = strings.ToUpper(i.FileName)

	if !i.fileNameSatisfiesISOConstraints(i.FileName) {
		return nil, fmt.Errorf("File name '%s' violates ISO9660 constraints", i.FileName)
	}

	buf := bytes.NewBuffer([]byte{})
	bufw := bufio.NewWriter(buf)
	w := NewISO9660Writer(bufw)

	err := i.writePrimaryVolumeDescriptor(w)
	if err != nil {
		return nil, err
	}

	err = i.writeVolumeDescriptorSetTerminator(w)
	if err != nil {
		return nil, err
	}

	i.writePathTable(w, binary.LittleEndian)
	i.writePathTable(w, binary.BigEndian)

	err = i.writeData(w)
	if err != nil {
		return nil, err
	}

	w.Finish()
	bufw.Flush()

	reservedBytes := make([]byte, int64(16*SectorSize))

	return append(reservedBytes, buf.Bytes()...), nil
}

func (i ISO9660) writePrimaryVolumeDescriptor(w *ISO9660Writer) error {
	if len(i.FileName) > 32 {
		i.FileName = i.FileName[:32]
	}
	now := time.Now()

	sw, err := w.NextSector()
	if err != nil {
		return err
	}
	if w.CurrentSector() != primaryVolumeSectorNum {
		return fmt.Errorf("internal error: unexpected primary volume sector %d", w.CurrentSector())
	}

	sw.WriteByte('\x01')
	sw.WriteString(volumeDescriptorSetMagic)
	sw.WriteByte('\x00')

	sw.WritePaddedString("", 32)
	sw.WritePaddedString(i.FileName, 32)

	sw.WriteZeros(8)
	sw.WriteBothEndianDWord(i.numTotalSectors())
	sw.WriteZeros(32)

	sw.WriteBothEndianWord(1) // volume set size
	sw.WriteBothEndianWord(1) // volume sequence number
	sw.WriteBothEndianWord(uint16(SectorSize))
	sw.WriteBothEndianDWord(SectorSize) // path table length

	sw.WriteLittleEndianDWord(littleEndianPathTableSectorNum)
	sw.WriteLittleEndianDWord(0) // no secondary path tables
	sw.WriteBigEndianDWord(bigEndianPathTableSectorNum)
	sw.WriteBigEndianDWord(0) // no secondary path tables

	WriteDirectoryRecord(sw, "\x00", rootDirectorySectorNum) // root directory

	sw.WritePaddedString("", 128) // volume set identifier
	sw.WritePaddedString("", 128) // publisher identifier
	sw.WritePaddedString("", 128) // data preparer identifier
	sw.WritePaddedString("", 128) // application identifier

	sw.WritePaddedString("", 37) // copyright file identifier
	sw.WritePaddedString("", 37) // abstract file identifier
	sw.WritePaddedString("", 37) // bibliographical file identifier

	sw.WriteDateTime(now)         // volume creation
	sw.WriteDateTime(now)         // most recent modification
	sw.WriteUnspecifiedDateTime() // expires
	sw.WriteUnspecifiedDateTime() // is effective (?)

	sw.WriteByte('\x01') // version
	sw.WriteByte('\x00') // reserved

	sw.PadWithZeros() // 512 (reserved for app) + 653 (zeros)

	return nil
}

func (i ISO9660) writeVolumeDescriptorSetTerminator(w *ISO9660Writer) error {
	sw, err := w.NextSector()
	if err != nil {
		return err
	}
	if w.CurrentSector() != primaryVolumeSectorNum+1 {
		return fmt.Errorf("internal error: unexpected volume descriptor set terminator sector %d", w.CurrentSector())
	}

	sw.WriteByte('\xFF')
	sw.WriteString(volumeDescriptorSetMagic)

	sw.PadWithZeros()

	return nil
}

func (i ISO9660) writePathTable(w *ISO9660Writer, bo binary.ByteOrder) error {
	sw, err := w.NextSector()
	if err != nil {
		return err
	}
	sw.WriteByte(1) // name length
	sw.WriteByte(0) // number of sectors in extended attribute record
	sw.WriteDWord(bo, rootDirectorySectorNum)
	sw.WriteWord(bo, 1) // parent directory recno (root directory)
	sw.WriteByte(0)     // identifier (root directory)
	sw.WriteByte(1)     // padding
	sw.PadWithZeros()
	return nil
}

func (i ISO9660) writeData(w *ISO9660Writer) error {
	sw, err := w.NextSector()
	if err != nil {
		return err
	}
	if w.CurrentSector() != rootDirectorySectorNum {
		return fmt.Errorf("internal error: unexpected root directory sector %d", w.CurrentSector())
	}

	WriteDirectoryRecord(sw, "\x00", w.CurrentSector())
	WriteDirectoryRecord(sw, "\x01", rootDirectorySectorNum)
	WriteFileRecordHeader(sw, i.FileName, w.CurrentSector()+1, i.inputLen())

	inputBuf := bytes.NewBuffer(i.Contents)

	// Now stream the data.  Note that the first buffer is never of SectorSize,
	// since we've already filled a part of the sector.
	b := make([]byte, SectorSize)
	total := uint32(0)
	for {
		l, err := inputBuf.Read(b)
		if err != nil && err != io.EOF {
			return fmt.Errorf("could not read from input file: %s", err)
		}
		if l > 0 {
			sw, err := w.NextSector()
			if err != nil {
				return err
			}
			sw.Write(b[:l])
			total += uint32(l)
		}
		if err == io.EOF {
			break
		}
	}
	if total != i.inputLen() {
		return fmt.Errorf("input file size changed (expected to read %d, read %d)", i.inputLen(), total)
	} else if w.CurrentSector() != i.numTotalSectors()-1 {
		return fmt.Errorf("internal error: unexpected last sector number (expected %d, actual %d)", i.numTotalSectors()-1, w.CurrentSector())
	}

	return nil
}

func (i ISO9660) numTotalSectors() uint32 {
	var numDataSectors uint32
	numDataSectors = (i.inputLen() + (SectorSize - 1)) / SectorSize
	return 1 + rootDirectorySectorNum + numDataSectors
}

func (ISO9660) fileNameSatisfiesISOConstraints(filename string) bool {
	invalidCharacter := func(r rune) bool {
		// According to ISO9660, only capital letters, digits, and underscores
		// are permitted.  Some sources say a dot is allowed as well.  I'm too
		// lazy to figure it out right now.
		if r >= 'A' && r <= 'Z' {
			return false
		} else if r >= '0' && r <= '9' {
			return false
		} else if r == '_' {
			return false
		} else if r == '.' {
			return false
		}
		return true
	}
	return strings.IndexFunc(filename, invalidCharacter) == -1
}
