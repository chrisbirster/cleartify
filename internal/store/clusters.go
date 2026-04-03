package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sqldata "cleartify/data"

	"cleartify/internal/domain"
)

var listClustersStatement = sqldata.MustQuery("list_clusters.sql")
var lookupClusterStatement = sqldata.MustQuery("lookup_cluster.sql")
var listClusterSignalsStatement = sqldata.MustQuery("list_cluster_signals.sql")
var updateClusterStatusStatement = sqldata.MustQuery("update_cluster_status.sql")

func (s *Store) ListClusters(ctx context.Context) ([]domain.Cluster, error) {
	rows, err := s.DB.QueryContext(ctx, listClustersStatement)
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	defer rows.Close()

	var clusters []domain.Cluster
	for rows.Next() {
		cluster, err := s.scanCluster(ctx, rows)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, rows.Err()
}

func (s *Store) ResolveCluster(ctx context.Context, clusterID string) (domain.Cluster, error) {
	updatedAt := time.Now().UTC().Truncate(time.Second)
	result, err := s.DB.ExecContext(ctx, updateClusterStatusStatement, "resolved", updatedAt.Format(time.RFC3339), clusterID)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("resolve cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("resolve cluster rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Cluster{}, sql.ErrNoRows
	}

	return s.LookupCluster(ctx, clusterID)
}

func (s *Store) LookupCluster(ctx context.Context, clusterID string) (domain.Cluster, error) {
	row := s.DB.QueryRowContext(ctx, lookupClusterStatement, clusterID)
	cluster, err := s.scanCluster(ctx, row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Cluster{}, sql.ErrNoRows
		}
		return domain.Cluster{}, fmt.Errorf("lookup cluster: %w", err)
	}
	return cluster, nil
}

type clusterScanner interface {
	Scan(dest ...any) error
}

func (s *Store) scanCluster(ctx context.Context, scanner clusterScanner) (domain.Cluster, error) {
	var cluster domain.Cluster
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&cluster.ID,
		&cluster.Title,
		&cluster.Summary,
		&cluster.Severity,
		&cluster.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.Cluster{}, fmt.Errorf("scan cluster: %w", err)
	}

	var err error
	cluster.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("decode cluster created_at: %w", err)
	}
	cluster.UpdatedAt, err = decodeTime(updatedAt)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("decode cluster updated_at: %w", err)
	}
	cluster.Signals, err = s.listSignalsForCluster(ctx, cluster.ID)
	if err != nil {
		return domain.Cluster{}, err
	}

	return cluster, nil
}

func (s *Store) listSignalsForCluster(ctx context.Context, clusterID string) ([]domain.Signal, error) {
	rows, err := s.DB.QueryContext(ctx, listClusterSignalsStatement, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list cluster signals: %w", err)
	}
	defer rows.Close()

	var signals []domain.Signal
	for rows.Next() {
		var signal domain.Signal
		var rawMetadata sql.NullString
		var createdAt string
		if err := rows.Scan(
			&signal.ID,
			&signal.SourceType,
			&signal.SourceID,
			&signal.Title,
			&signal.Body,
			&signal.URL,
			&rawMetadata,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan cluster signal: %w", err)
		}
		signal.Metadata, err = decodeJSONMap(rawMetadata)
		if err != nil {
			return nil, fmt.Errorf("decode cluster signal metadata: %w", err)
		}
		signal.CreatedAt, err = decodeTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("decode cluster signal time: %w", err)
		}
		signals = append(signals, signal)
	}

	return signals, rows.Err()
}
