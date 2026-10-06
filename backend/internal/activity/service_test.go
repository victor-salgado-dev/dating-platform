package activity

import (
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSortItemsOrdersMixedEventsNewestFirst(t *testing.T) {
	older := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	newer := older.Add(time.Minute)
	items := []Item{
		{EventType: EventFavoriteReceived, ProfileID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), CreatedAt: older},
		{EventType: EventLikeReceived, ProfileID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), CreatedAt: newer},
		{EventType: EventMatchCreated, ProfileID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), CreatedAt: older},
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	if items[0].EventType != EventLikeReceived || items[1].EventType != EventFavoriteReceived || items[2].EventType != EventMatchCreated {
		t.Fatalf("orden de actividad = %#v", items)
	}
}
