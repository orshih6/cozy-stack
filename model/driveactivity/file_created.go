package driveactivity

import (
	"context"

	"github.com/cozy/cozy-stack/model/instance"
	"github.com/cozy/cozy-stack/model/vfs"
	"github.com/cozy/cozy-stack/pkg/messaging"
	"github.com/cozy/cozy-stack/pkg/rabbitmq"
)

// FileCreatedData is the data of a com.twake.drive.file.created.v1 event.
type FileCreatedData struct {
	Object     FileObject  `json:"object"`
	Recipients []Recipient `json:"recipients"`
}

// FileObject is the created file: its id, name and folder, never its content.
type FileObject struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Folder    Reference `json:"folder"`
	Container Container `json:"container"`
}

// Reference is the id of something an event points to.
type Reference struct {
	ID string `json:"id"`
}

// Recipient is a person an event is for, by uuid.
type Recipient struct {
	UUID   string `json:"uuid"`
	Reason string `json:"reason"`
}

// FileEventsOn says if the context of the instance publishes an event per
// file created in the personal drive. It is off by default, as a sync can
// create thousands of files at once (see ADR 010).
func FileEventsOn(inst *instance.Instance) bool {
	settings, ok := inst.SettingsContext()
	if !ok {
		return false
	}
	cfg, _ := settings["drive_activity"].(map[string]interface{})
	on, _ := cfg["file_created"].(bool)
	return on
}

// FileCreated publishes that a file was created in the personal drive of the
// instance. byOwner tells if the owner created it, rather than someone with a
// sharing link. The file is already created, so the event is best effort:
// FileCreated only logs its errors.
func FileCreated(ctx context.Context, pub messaging.Publisher, inst *instance.Instance, file *vfs.FileDoc, byOwner bool) {
	if pub == nil || file == nil || !FileEventsOn(inst) {
		return
	}
	// The events name people by uuid only: without it, no one could get it.
	if inst.OIDCUUID == "" {
		inst.Logger().WithNamespace("drive-activity").
			Debugf("No file event: the uuid of the owner is unknown until they log in")
		return
	}
	event := rabbitmq.CloudEvent{
		ID:       file.DocID,
		Type:     rabbitmq.RoutingKeyDriveFileCreated,
		Time:     file.CreatedAt.UTC(),
		TwakeOrg: inst.OrgID,
		Data: FileCreatedData{
			Object: FileObject{
				Type:      "file",
				ID:        file.DocID,
				Title:     file.DocName,
				Folder:    Reference{ID: file.DirID},
				Container: Container{Kind: "drive", ID: inst.Domain},
			},
			Recipients: []Recipient{{UUID: inst.OIDCUUID, Reason: "owner"}},
		},
	}
	if byOwner {
		event.TwakeActorID = inst.OIDCUUID
	}
	if err := pub.Publish(ctx, rabbitmq.ActivityRequest(inst.ContextName, event)); err != nil {
		inst.Logger().WithNamespace("drive-activity").
			Warnf("Cannot publish the creation of file %s: %s", file.DocID, err)
	}
}
