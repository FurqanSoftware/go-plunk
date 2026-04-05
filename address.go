package plunk

import "encoding/json"

// Address represents an email address with an optional display name.
//
// When marshaled to JSON, an Address without a Name is encoded as a plain
// string. An Address with a Name is encoded as {"name": "...", "email": "..."}.
type Address struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// Addr returns an Address with only an email and no display name.
func Addr(email string) Address {
	return Address{Email: email}
}

func (a Address) MarshalJSON() ([]byte, error) {
	if a.Name == "" {
		return json.Marshal(a.Email)
	}
	return json.Marshal(struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{a.Name, a.Email})
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		a.Email = s
		return nil
	}
	var obj struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	a.Name = obj.Name
	a.Email = obj.Email
	return nil
}

type toField []Address

func (t toField) MarshalJSON() ([]byte, error) {
	if len(t) == 1 {
		return json.Marshal(t[0])
	}
	type raw []Address
	return json.Marshal(raw(t))
}
