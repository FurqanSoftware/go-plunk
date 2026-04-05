package plunk

import (
	"encoding/json"
	"testing"
)

func TestAddressMarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		addr Address
		want string
	}{
		{"email only", Addr("user@example.com"), `"user@example.com"`},
		{"with name", Address{Name: "Alice", Email: "alice@example.com"}, `{"name":"Alice","email":"alice@example.com"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.addr)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAddressUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantEmail string
	}{
		{"string", `"user@example.com"`, "", "user@example.com"},
		{"object", `{"name":"Alice","email":"alice@example.com"}`, "Alice", "alice@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addr Address
			if err := json.Unmarshal([]byte(tt.input), &addr); err != nil {
				t.Fatal(err)
			}
			if addr.Name != tt.wantName || addr.Email != tt.wantEmail {
				t.Errorf("got {%q, %q}, want {%q, %q}", addr.Name, addr.Email, tt.wantName, tt.wantEmail)
			}
		})
	}
}

func TestToFieldMarshalJSON(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		got, err := json.Marshal(toField{Addr("a@b.com")})
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != `"a@b.com"` {
			t.Errorf("got %s, want %q", got, "a@b.com")
		}
	})

	t.Run("multiple", func(t *testing.T) {
		got, err := json.Marshal(toField{Addr("a@b.com"), Address{Name: "B", Email: "b@c.com"}})
		if err != nil {
			t.Fatal(err)
		}
		want := `["a@b.com",{"name":"B","email":"b@c.com"}]`
		if string(got) != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}
