package server

import (
	gliderssh "github.com/gliderlabs/ssh"
)

func (s *Server) passwordHandler(ctx gliderssh.Context, pass string) bool {
	return s.authenticator.Password(ctx, ctx.User(), pass)
}

func (s *Server) publicKeyHandler(ctx gliderssh.Context, key gliderssh.PublicKey) bool {
	return s.authenticator.PublicKey(ctx, ctx.User(), key)
}
