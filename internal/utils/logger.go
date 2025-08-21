package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// Custom formatter to maintain exact field order
type OrderedJSONFormatter struct {
	TimestampFormat string
}

func (f *OrderedJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{})
	
	// Maintain exact order: date, level, source, description, meta
	orderedData := []string{"date", "level", "source", "description"}
	
	// Add fields in the correct order
	data["date"] = entry.Time.Format(f.TimestampFormat)
	data["level"] = strings.ToUpper(entry.Level.String())
	data["source"] = "aken-reporting"
	data["description"] = entry.Message
	
	// Add meta if it exists
	if meta, exists := entry.Data["meta"]; exists {
		data["meta"] = meta
		orderedData = append(orderedData, "meta")
	}
	
	// Add any other fields (shouldn't normally happen with our usage)
	for k, v := range entry.Data {
		if k != "meta" && k != "source" {
			data[k] = v
		}
	}
	
	// Build ordered JSON manually to maintain field order
	var result strings.Builder
	result.WriteString("{")
	
	for i, key := range orderedData {
		if i > 0 {
			result.WriteString(",")
		}
		keyBytes, _ := json.Marshal(key)
		valueBytes, _ := json.Marshal(data[key])
		result.WriteString(fmt.Sprintf("%s:%s", keyBytes, valueBytes))
	}
	
	result.WriteString("}\n")
	return []byte(result.String()), nil
}

func init() {
	Logger = logrus.New()
	
	// Use custom formatter to maintain field order
	Logger.SetFormatter(&OrderedJSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	
	// Set log level (can be configured via environment)
	Logger.SetLevel(logrus.TraceLevel)
}

// Helper functions to log with consistent format
func LogHTTPRequest(method, url, status string, responseTime string, requestID string) {
	Logger.WithFields(logrus.Fields{
		"source": "aken-reporting",
		"meta": map[string]interface{}{
			"date":          time.Now().Format(time.RFC3339Nano),
			"method":        method,
			"url":           url,
			"status":        status,
			"id":            requestID,
			"response_time": responseTime,
		},
	}).Trace("HTTP Request")
}

func LogTransactionRequest(merchantID, sessionID string, operation string, additionalFields map[string]interface{}) {
	fields := logrus.Fields{
		"source": "aken-reporting",
	}
	
	// Add meta fields
	meta := map[string]interface{}{}
	if merchantID != "" {
		meta["merchant_id"] = merchantID
	}
	if sessionID != "" {
		meta["session_id"] = sessionID
	}
	if operation != "" {
		meta["operation"] = operation
	}
	
	// Add any additional fields to meta
	for key, value := range additionalFields {
		meta[key] = value
	}
	
	fields["meta"] = meta
	
	Logger.WithFields(fields).Trace("Transaction request received")
}

func LogTransactionResponse(merchantID, sessionID string, success bool, result string, additionalFields map[string]interface{}) {
	fields := logrus.Fields{
		"source": "aken-reporting",
	}
	
	// Add meta fields
	meta := map[string]interface{}{}
	if merchantID != "" {
		meta["merchant_id"] = merchantID
	}
	if sessionID != "" {
		meta["session_id"] = sessionID
	}
	meta["success"] = success
	meta["result"] = result
	
	// Add any additional fields to meta
	for key, value := range additionalFields {
		meta[key] = value
	}
	
	fields["meta"] = meta
	
	Logger.WithFields(fields).Trace("Transaction response")
}

func LogInfo(message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"source": "aken-reporting",
	}
	
	if len(fields) > 0 {
		logFields["meta"] = fields
	}
	
	Logger.WithFields(logFields).Info(message)
}

func LogTrace(message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"source": "aken-reporting",
	}
	
	if len(fields) > 0 {
		logFields["meta"] = fields
	}
	
	Logger.WithFields(logFields).Trace(message)
}

func LogError(message string, err error, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"source": "aken-reporting",
	}
	
	meta := map[string]interface{}{}
	if err != nil {
		meta["error"] = err.Error()
	}
	
	// Add any additional fields to meta
	for key, value := range fields {
		meta[key] = value
	}
	
	if len(meta) > 0 {
		logFields["meta"] = meta
	}
	
	Logger.WithFields(logFields).Error(message)
}

func LogWarn(message string, fields map[string]interface{}) {
	logFields := logrus.Fields{
		"source": "aken-reporting",
	}
	
	if len(fields) > 0 {
		logFields["meta"] = fields
	}
	
	Logger.WithFields(logFields).Warn(message)
}