package types

import "os"

type (
	ClientStatusType string
	UserStatusType   string
)

// GroupAdmin is the name of the admin user, defaults to "admin"
// if the env ADMIN_USERNAME is not set
var (
	GroupAdmin = func() string {
		username := os.Getenv("ADMIN_USERNAME")
		if username == "" {
			username = "admin"
		}
		return username
	}()

	GroupAuditor = func() string {
		auditor := os.Getenv("AUDITOR_USERNAME")
		if auditor == "" {
			return "auditor"
		}
		return auditor
	}()
)

const (
	// ClientStatusReady indicates the grpc client is ready to
	// subscribe to a new connection
	ClientStatusReady ClientStatusType = "ready"
	// ClientStatusConnected indicates the client has opened a new session
	ClientStatusConnected ClientStatusType = "connected"
	// ClientStatusDisconnected indicates the grpc client has disconnected
	ClientStatusDisconnected ClientStatusType = "disconnected"

	UserStatusActive    UserStatusType = "active"
	UserStatusReviewing UserStatusType = "reviewing"
	UserStatusInactive  UserStatusType = "inactive"
)

type ReviewStatus string

const (
	ReviewStatusPending    ReviewStatus = "PENDING"
	ReviewStatusApproved   ReviewStatus = "APPROVED"
	ReviewStatusRejected   ReviewStatus = "REJECTED"
	ReviewStatusRevoked    ReviewStatus = "REVOKED"
	ReviewStatusProcessing ReviewStatus = "PROCESSING"
	ReviewStatusExecuted   ReviewStatus = "EXECUTED"
	ReviewStatusUnknown    ReviewStatus = "UNKNOWN"
)

const (
	SessionStatusOpen  = "open"
	SessionStatusReady = "ready"
	SessionStatusDone  = "done"
)
