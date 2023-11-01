package main

type Metrics struct {
	MType string `json:"type"` // параметр, принимающий значение gauge или counter
}

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)

func main() {
	m := Metrics{
		MType: "counter",
	}

	switch m.MType {
	case CounterType:
		println("counter")
	case GaugeType:
		println("gauge")
	default:
		println("default")
	}
}
