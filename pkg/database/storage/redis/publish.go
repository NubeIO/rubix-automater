package redis

const (
	TopicJob = "job"
)

// Pub posts the message to the channel.
func (rs *Redis) Pub(channel string, message interface{}) error {
	err := rs.pub.Publish(channel, message)
	if err != nil {
		return err
	}
	return nil
}
