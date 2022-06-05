package redis

import "encoding/json"

const (
	TopicJob = "job"
)

// Pub posts the message to the channel.
func (rs *Redis) Pub(channel string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = rs.Publish(ctx, channel, payload).Err()
	if err != nil {
		return err
	}
	return nil
}
