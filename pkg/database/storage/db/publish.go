package redis

const (
	TopicJob = "job"
)

// Pub posts the message to the channel.
func (inst *Redis) Pub(channel string, message interface{}) error {
	err := inst.pub.Publish(channel, message)
	if err != nil {
		return err
	}
	return nil
}
