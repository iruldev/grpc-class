package repository

import "sync"

type Rating struct {
	Count uint32
	Sum   float64
}

type RatingRepository interface {
	Add(laptopID string, score float64) (*Rating, error)
}

type RatingRepositoryImpl struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

func NewRatingRepository() RatingRepository {
	return &RatingRepositoryImpl{rating: make(map[string]*Rating)}
}

func (r *RatingRepositoryImpl) Add(laptopID string, score float64) (*Rating, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	rating := r.rating[laptopID]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	r.rating[laptopID] = rating
	return rating, nil
}
