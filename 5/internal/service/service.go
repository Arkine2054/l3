package service

import (
	"context"
	"log"
	"sync"
	"time"

	"gitlab.com/arkine/l3/5/internal/models"
	"gitlab.com/arkine/l3/5/internal/repository"
)

// Service инкапсулирует бизнес-логику приложения
type Service struct {
	repo         *repository.Repository
	expireAfter  time.Duration
	cancelTicker *time.Ticker
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewService создаёт новый сервис
func NewService(repo *repository.Repository, expireAfter time.Duration) *Service {
	return &Service{
		repo:        repo,
		expireAfter: expireAfter,
		stopChan:    make(chan struct{}),
	}
}

// CreateEvent создаёт новое мероприятие
func (s *Service) CreateEvent(ctx context.Context, title string, date time.Time, totalSeats int) (*models.Event, error) {
	event := &models.Event{
		Title:      title,
		Date:       date,
		TotalSeats: totalSeats,
	}
	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *Service) ListEvents(ctx context.Context) ([]models.EventInfo, error) {
	return s.repo.ListEvents(ctx)
}

// GetEvent возвращает мероприятие по ID с подсчётом свободных мест
func (s *Service) GetEvent(ctx context.Context, id int) (*models.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	booked, err := s.repo.CountBookedSeats(ctx, id)
	if err != nil {
		return nil, err
	}
	event.AvailableSeats = event.TotalSeats - booked
	if event.AvailableSeats < 0 {
		event.AvailableSeats = 0
	}
	return event, nil
}

// BookSeat бронирует место на мероприятие
func (s *Service) BookSeat(ctx context.Context, eventID int, userName string) (*models.Booking, error) {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	booked, err := s.repo.CountBookedSeats(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if booked >= event.TotalSeats {
		return nil, repository.ErrNoSeatsAvailable
	}
	return s.repo.CreateBooking(ctx, eventID, userName)
}

// ConfirmBooking подтверждает оплату брони
func (s *Service) ConfirmBooking(ctx context.Context, bookingID int) error {
	return s.repo.MarkBookingPaid(ctx, bookingID)
}

// CancelExpiredBookings удаляет неоплаченные брони старше заданного интервала
func (s *Service) CancelExpiredBookings(ctx context.Context) error {
	expirationTime := time.Now().Add(-s.expireAfter)
	log.Printf("[worker] отменяем неоплаченные брони, старше: %s", expirationTime.Format(time.RFC3339))
	return s.repo.DeleteExpiredBookings(ctx, expirationTime)
}

// StartBookingCleaner запускает фоновый процесс очистки просроченных броней
func (s *Service) StartBookingCleaner(interval time.Duration) {
	if s.cancelTicker != nil {
		log.Println("Booking cleaner уже запущен")
		return
	}

	s.cancelTicker = time.NewTicker(interval)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.cancelTicker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := s.CancelExpiredBookings(ctx); err != nil {
					log.Printf("Ошибка очистки броней: %v", err)
				}
				cancel()

			case <-s.stopChan:
				log.Println("Остановка фонового очистителя броней")
				return
			}
		}
	}()
}

// StopBookingCleaner останавливает фоновую задачу очистки
func (s *Service) StopBookingCleaner() {
	if s.cancelTicker == nil {
		return
	}
	s.cancelTicker.Stop()
	close(s.stopChan)
	s.wg.Wait()
	s.cancelTicker = nil
	log.Println("Фоновая очистка броней остановлена")
}
