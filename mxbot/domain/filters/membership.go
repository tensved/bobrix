package filters

import "maunium.net/go/mautrix/event"

// FilterMembershipEvent - filter for membership events
// check if message type is event.Membership
func FilterMembershipEvent(m event.Membership) Filter {
	return func(evt *event.Event) bool {
		return evt.Content.AsMember().Membership == m
	}
}

// FilterMembershipInvite - filter for invite messages
// check if message type is event.MembershipInvite
func FilterMembershipInvite() Filter {
	return FilterMembershipEvent(event.MembershipInvite)
}
