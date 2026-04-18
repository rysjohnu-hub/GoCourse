package main

import "testing"

func TestAddBasic(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d; want %d", got, want)
	}
}

func TestSubtractTableDriven(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"both positive", 10, 3, 7},
		{"positive minus zero", 5, 0, 5},
		{"negative minus positive", -1, 4, -5},
		{"both negative", -2, -3, 1},
		{"zero minus positive", 0, 5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Subtract(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivideTableDriven(t *testing.T) {
	tests := []struct {
		name      string
		a, b      int
		want      int
		wantError bool
	}{
		{"normal division", 10, 2, 5, false},
		{"division with remainder", 7, 2, 3, false},
		{"zero divided by number", 0, 5, 0, false},
		{"negative divided by positive", -10, 2, -5, false},
		{"division by zero", 10, 0, 0, true},
		{"negative divided by negative", -10, -2, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)

			if (err != nil) != tt.wantError {
				t.Errorf("Divide(%d, %d) error = %v; wantError %v", tt.a, tt.b, err, tt.wantError)
				return
			}

			if !tt.wantError && got != tt.want {
				t.Errorf("Divide(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
