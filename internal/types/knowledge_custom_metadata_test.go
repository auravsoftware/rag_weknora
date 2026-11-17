package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKnowledgeCustomMetadataTextIsStableAndSeparate(t *testing.T) {
	knowledge := &Knowledge{
		Metadata:       JSON(`{"external_id":"secret-internal-id"}`),
		CustomMetadata: JSON(`{"region":"Shanghai","department":"R&D"}`),
	}
	require.Equal(t, "department: R&D\nregion: Shanghai", knowledge.CustomMetadataText())
	require.NotContains(t, knowledge.CustomMetadataText(), "external_id")
}

func TestKnowledgeCustomMetadataVietnameseFieldsAndValidity(t *testing.T) {
	knowledge := &Knowledge{
		CustomMetadata: JSON(`{"document_title":"Quyết định 01/2025","issuing_authority":"Bộ Tư pháp","issue_date":"2025-01-10","effective_date":"2025-01-15","expiration_date":"2099-12-31"}`),
	}
	text := knowledge.CustomMetadataText()
	require.Contains(t, text, "Tiêu đề văn bản (document_title): Quyết định 01/2025")
	require.Contains(t, text, "Nơi ban hành (issuing_authority): Bộ Tư pháp")
	require.Contains(t, text, "Ngày ban hành (issue_date): 2025-01-10")
	require.Contains(t, text, "Ngày hiệu lực (effective_date): 2025-01-15")
	require.Contains(t, text, "Ngày hết hiệu lực (expiration_date): 2099-12-31")
	require.Contains(t, text, "Trạng thái hiệu lực: Đang có hiệu lực")
}

