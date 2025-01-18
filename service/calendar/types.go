package calendar

type CalendarService interface {
	CreateCalendar(userEmail string) error
	PlanOrLog(userEmail string, planOrLog string) error
	GetCalendars(userEmail string) ([]string, error)
}