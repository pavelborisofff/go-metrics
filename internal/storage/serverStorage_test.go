package storage

import "testing"

func TestServerStorage_UpdateGauge(t *testing.T) {
	type testType struct {
		name          string
		value         Gauge
		expectedValue Gauge
	}

	tests := []testType{
		{
			name:          "UpdateGauge",
			value:         Gauge(1),
			expectedValue: Gauge(1),
		},
		{
			name:          "UpdateGauge Big Number",
			value:         Gauge(1234567890.123456789),
			expectedValue: Gauge(1234567890.123456789),
		},
		{
			name:          "UpdateGauge Negative Number",
			value:         Gauge(-1),
			expectedValue: Gauge(-1),
		},
		{
			name:          "UpdateGauge Zero",
			value:         Gauge(0),
			expectedValue: Gauge(0),
		},
		{
			name:          "UpdateGauge Big Negative Number",
			value:         Gauge(-1234567890.123456789),
			expectedValue: Gauge(-1234567890.123456789),
		},
	}

	for _, test := range tests {
		s := NewMemStorage()
		s.UpdateGauge(test.name, test.value)
		if s.GaugeStorage[test.name] != test.expectedValue {
			t.Errorf("UpdateGauge() = %v, want %v", s.GaugeStorage[test.name], test.expectedValue)
		}
	}
}

func TestServerStorage_IncrementCounter(t *testing.T) {
	type testType struct {
		name         string
		defaultValue Counter
		addValue     Counter
		expectedVal  Counter
	}

	tests := []testType{
		{
			name:         "IncrementCounter",
			defaultValue: Counter(1),
			addValue:     Counter(1),
			expectedVal:  Counter(2),
		},
		{
			name:         "IncrementCounter Big Number",
			defaultValue: Counter(0),
			addValue:     Counter(1234567890),
			expectedVal:  Counter(1234567890),
		},
		{
			name:         "IncrementCounter Zero",
			defaultValue: Counter(1),
			addValue:     Counter(0),
			expectedVal:  Counter(1),
		},
	}

	for _, test := range tests {
		s := NewMemStorage()

		s.CounterStorage[test.name] = test.defaultValue
		s.IncrementCounter(test.name, test.addValue)

		if s.CounterStorage[test.name] != test.expectedVal {
			t.Errorf("IncrementCounter() = %v, want %v", s.CounterStorage[test.name], test.expectedVal)
		}
	}
}
