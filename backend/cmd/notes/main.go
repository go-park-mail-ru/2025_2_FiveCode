package notes

// TODO: make crsf protection for notes service

// func (u *AuthUsecase) GenerateCSRFToken(ctx context.Context, sessionID string) (string, error) {
// 	log := logger.FromContext(ctx)
// 	log.Info().Str("session_id", sessionID).Msg("generating csrf token")

// 	token, err := apiutils.GenerateCSRFToken(sessionID, u.CSRFSecret)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed to generate csrf token")
// 		return "", fmt.Errorf("failed to generate csrf token: %w", err)
// 	}

// 	return token, nil
// }