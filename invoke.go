package ldi

import (
	"log"
)

func (d *Di) MustInvoke(functions ...any) *Di {
	if err := d.Invoke(functions...); err != nil {
		log.Fatal(err)
	}
	return d
}

func (d *Di) Invoke(functions ...any) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, function := range functions {
		_, err := d.invoke(function)
		if err != nil {
			return err
		}
	}
	return nil
}
