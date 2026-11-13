package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_ParsePublicID(t *testing.T) {
	id := uuid.MustParse("01900000-0000-7000-8000-000000000001")

	tests := []struct {
		name     string
		prefix   string
		publicID string
		want     uuid.UUID
		wantErr  bool
	}{
		{name: "valid", prefix: PrefixDeliveryID, publicID: "del_" + id.String(), want: id},
		{name: "round trip", prefix: PrefixEndpointID, publicID: ToPublicEndpointID(id), want: id},
		{name: "wrong prefix", prefix: PrefixDeliveryID, publicID: "ep_" + id.String(), wantErr: true},
		{name: "prefix without separator", prefix: PrefixDeliveryID, publicID: "del" + id.String(), wantErr: true},
		{name: "no prefix", prefix: PrefixDeliveryID, publicID: id.String(), wantErr: true},
		{name: "malformed uuid", prefix: PrefixDeliveryID, publicID: "del_123", wantErr: true},
		{name: "empty uuid", prefix: PrefixDeliveryID, publicID: "del_", wantErr: true},
		{name: "empty string", prefix: PrefixDeliveryID, publicID: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePublicID(tt.prefix, tt.publicID)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidID)
				require.Equal(t, uuid.Nil, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
