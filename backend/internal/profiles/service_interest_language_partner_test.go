func (f *fakeRepository) GetPublicPhoto(ctx context.Context, profileID, photoID, viewerUserID uuid.UUID) (*Photo, error) {
    return nil, ErrPhotoNotFound
}
