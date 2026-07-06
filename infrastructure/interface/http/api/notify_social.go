package httpapi

import (
	"context"
	"log"
)

// notifyPostLiked writes an in-app notification to the post author (sync, detached context).
func (s *Server) notifyPostLiked(postID, likerID int64) {
	if s.Notifications == nil || s.Posts == nil {
		return
	}
	ctx := context.Background()
	post, err := s.Posts.GetByID(ctx, postID)
	if err != nil || post == nil || post.UserID == likerID {
		return
	}
	if err := s.Notifications.OnPostLiked(ctx, post.UserID, postID, likerID); err != nil {
		log.Printf("notify post liked: %v", err)
	}
}

// notifyUserFollowed writes an in-app notification to the followee (sync, detached context).
func (s *Server) notifyUserFollowed(followeeID, followerID int64) {
	if s.Notifications == nil || followeeID == 0 || followerID == 0 || followeeID == followerID {
		return
	}
	if err := s.Notifications.OnUserFollowed(context.Background(), followeeID, followerID); err != nil {
		log.Printf("notify user followed: %v", err)
	}
}

// notifyReplyLiked writes an in-app notification to the reply author (sync, detached context).
func (s *Server) notifyReplyLiked(replyID, likerID int64) {
	if s.Notifications == nil || s.Replies == nil {
		return
	}
	ctx := context.Background()
	rep, err := s.Replies.GetByID(ctx, replyID)
	if err != nil || rep == nil || rep.UserID == likerID {
		return
	}
	if err := s.Notifications.OnReplyLiked(ctx, rep.UserID, rep.PostID, replyID, likerID); err != nil {
		log.Printf("notify reply liked: %v", err)
	}
}
