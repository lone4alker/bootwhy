package parse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"time"
)

type Entry struct {
	Mono             time.Duration
	Unit, MsgID, Msg string
	Priority         int
}

type journalRecord struct {
	Mono     json.RawMessage `json:"__MONOTONIC_TIMESTAMP"`
	Unit     string          `json:"UNIT"`
	Systemd  string          `json:"_SYSTEMD_UNIT"`
	MsgID    string          `json:"MESSAGE_ID"`
	Message  json.RawMessage `json:"MESSAGE"`
	Priority json.RawMessage `json:"PRIORITY"`
}

// ParseJournal parses newline-delimited journal JSON. Malformed records are
// counted and skipped; the cutoff stops parsing once the boot window is passed.
func ParseJournal(reader io.Reader, until time.Duration) ([]Entry, int, error) {
	var entries []Entry
	var malformed int
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		var record journalRecord
		if err := json.NewDecoder(bytes.NewReader(scanner.Bytes())).Decode(&record); err != nil {
			malformed++
			continue
		}
		entry, ok := decodeJournalRecord(record)
		if !ok {
			malformed++
			continue
		}
		if until > 0 && entry.Mono > until {
			break
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return entries, malformed, err
	}
	return entries, malformed, nil
}

func decodeJournalRecord(record journalRecord) (Entry, bool) {
	mono, err := decodeJournalInt(record.Mono)
	if err != nil {
		return Entry{}, false
	}
	message, err := decodeJournalMessage(record.Message)
	if err != nil {
		return Entry{}, false
	}
	priority, err := decodeJournalInt(record.Priority)
	if err != nil {
		priority = 0
	}
	unit := record.Unit
	if unit == "" {
		unit = record.Systemd
	}
	return Entry{
		Mono:     time.Duration(mono) * time.Microsecond,
		Unit:     unit,
		MsgID:    record.MsgID,
		Msg:      message,
		Priority: int(priority),
	}, true
}

func decodeJournalInt(value json.RawMessage) (int64, error) {
	var number json.Number
	if err := json.Unmarshal(value, &number); err == nil {
		return number.Int64()
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return 0, err
	}
	return strconv.ParseInt(text, 10, 64)
}

func decodeJournalMessage(value json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return text, nil
	}
	var data []byte
	if err := json.Unmarshal(value, &data); err != nil {
		return "", err
	}
	return string(data), nil
}
