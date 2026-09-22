package visits

import (
"time"

"github.com/google/uuid"

"dating-platform/backend/internal/profiles"
)

const DefaultPageSize = 20

type ListItem struct {
ProfileID        uuid.UUID
DisplayName      string
Age              int
Gender           profiles.Gender
CountryCode      string
Region           *string
RelationshipGoal *profiles.RelationshipGoal
HasPhoto         bool
VisitedAt        time.Time
}

type ListResult struct {
Items      []ListItem
Total      int
Page       int
PageSize   int
TotalPages int
}
