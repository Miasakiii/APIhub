package ccswitch

// ProviderEndpoint represents a single endpoint from the provider_endpoints table.
type ProviderEndpoint struct {
	ID         string
	ProviderID string
	AppType    string
	URL        string
}

// LookupEndpoints returns all endpoints for a given provider from provider_endpoints table.
// Falls back to settings_config parsing if the dedicated table has no entries.
func (r *Reader) LookupEndpoints(providerID, appType string) ([]ProviderEndpoint, error) {
	rows, err := r.db.Query(`
		SELECT id, provider_id, app_type, url
		FROM provider_endpoints
		WHERE provider_id = ? AND app_type = ?
	`, providerID, appType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eps []ProviderEndpoint
	for rows.Next() {
		var e ProviderEndpoint
		if err := rows.Scan(&e.ID, &e.ProviderID, &e.AppType, &e.URL); err != nil {
			return nil, err
		}
		eps = append(eps, e)
	}
	return eps, rows.Err()
}

// FetchAllEndpoints returns all endpoints, useful for bulk caching.
func (r *Reader) FetchAllEndpoints() ([]ProviderEndpoint, error) {
	rows, err := r.db.Query(`
		SELECT id, provider_id, app_type, url
		FROM provider_endpoints
		ORDER BY app_type, provider_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eps []ProviderEndpoint
	for rows.Next() {
		var e ProviderEndpoint
		if err := rows.Scan(&e.ID, &e.ProviderID, &e.AppType, &e.URL); err != nil {
			return nil, err
		}
		eps = append(eps, e)
	}
	return eps, rows.Err()
}
