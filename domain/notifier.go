package domain

type Notifier interface {
	Notify([]Notification) error
}

type Notification struct {
	User    string
	Message string
}
