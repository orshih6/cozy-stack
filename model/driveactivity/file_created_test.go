package driveactivity

import (
	"context"
	"testing"
	"time"

	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/model/vfs"
	"github.com/cozy/cozy-stack/pkg/config/config"
	"github.com/cozy/cozy-stack/pkg/rabbitmq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ownerUUID = "6208ad54-5683-1041-8c8f-c9e6dccfd4ac"

func fileEventsContext(t *testing.T, on bool) {
	t.Helper()
	contexts := config.GetConfig().Contexts
	t.Cleanup(func() { config.GetConfig().Contexts = contexts })
	config.GetConfig().Contexts = map[string]interface{}{
		config.DefaultInstanceContext: map[string]interface{}{
			"drive_activity": map[string]interface{}{"file_created": on},
		},
	}
}

func TestFileCreated(t *testing.T) {
	config.UseTestFile(t)
	ctx := context.Background()
	created := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	file := &vfs.FileDoc{DocID: "file-1", DirID: "dir-1", DocName: "invoice.pdf", CreatedAt: created}
	owner := func() *instance.Instance {
		return &instance.Instance{
			Domain:      "alice.twake.local",
			OrgID:       "acme",
			ContextName: config.DefaultInstanceContext,
			OIDCUUID:    ownerUUID,
		}
	}

	t.Run("IsOffByDefault", func(t *testing.T) {
		pub := &fakePublisher{}
		FileCreated(ctx, pub, owner(), file, true)
		assert.Empty(t, pub.requests)
	})

	t.Run("PublishesTheFileForItsOwner", func(t *testing.T) {
		fileEventsContext(t, true)
		pub := &fakePublisher{}
		FileCreated(ctx, pub, owner(), file, true)
		require.Len(t, pub.requests, 1)
		req := pub.requests[0]
		assert.Equal(t, rabbitmq.ExchangeActivity, req.Exchange)
		assert.Equal(t, rabbitmq.RoutingKeyDriveFileCreated, req.RoutingKey)
		assert.Equal(t, "file-1", req.MessageID)
		event := req.Payload.(rabbitmq.CloudEvent)
		assert.Equal(t, "twake://drive", event.Source)
		assert.Equal(t, "acme", event.TwakeOrg)
		assert.Equal(t, ownerUUID, event.TwakeActorID)
		assert.Equal(t, created, event.Time)
		assert.Equal(t, FileCreatedData{
			Object: FileObject{
				Type:      "file",
				ID:        "file-1",
				Title:     "invoice.pdf",
				Folder:    Reference{ID: "dir-1"},
				Container: Container{Kind: "drive", ID: "alice.twake.local"},
			},
			Recipients: []Recipient{{UUID: ownerUUID, Reason: "owner"}},
		}, event.Data)
	})

	t.Run("NamesNoActorForAFileFromASharingLink", func(t *testing.T) {
		fileEventsContext(t, true)
		pub := &fakePublisher{}
		FileCreated(ctx, pub, owner(), file, false)
		require.Len(t, pub.requests, 1)
		event := pub.requests[0].Payload.(rabbitmq.CloudEvent)
		assert.Empty(t, event.TwakeActorID)
		assert.Equal(t, []Recipient{{UUID: ownerUUID, Reason: "owner"}}, event.Data.(FileCreatedData).Recipients)
	})

	t.Run("PublishesNothingUntilTheUUIDOfTheOwnerIsKnown", func(t *testing.T) {
		fileEventsContext(t, true)
		pub := &fakePublisher{}
		inst := owner()
		inst.OIDCUUID = ""
		FileCreated(ctx, pub, inst, file, true)
		assert.Empty(t, pub.requests)
	})

	t.Run("OnlyLogsAFailedPublish", func(t *testing.T) {
		fileEventsContext(t, true)
		FileCreated(ctx, &fakePublisher{err: assert.AnError}, owner(), file, true)
		FileCreated(ctx, nil, owner(), file, true)
	})
}
