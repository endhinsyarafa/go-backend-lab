-- ============================================================================
-- 1. Rekap jam penggunaan per ruangan untuk bulan berjalan
--    Hanya menghitung booking berstatus CONFIRMED (CANCELLED tidak dihitung)
-- ============================================================================
SELECT
    r.id AS room_id,
    r.name,
    COALESCE(
        ROUND(
            SUM(EXTRACT(EPOCH FROM (b.end_time - b.start_time)) / 3600.0)::numeric,
            2
        ),
        0
    ) AS total_hours
FROM rooms r
LEFT JOIN bookings b
    ON b.room_id = r.id
    AND b.status = 'CONFIRMED'
    AND DATE_TRUNC('month', b.start_time) = DATE_TRUNC('month', CURRENT_TIMESTAMP)
GROUP BY r.id, r.name
ORDER BY r.id ASC;

-- ============================================================================
-- 2. Pencarian ruangan tersedia pada interval waktu tertentu
--    Hanya ruangan aktif tanpa booking CONFIRMED yang beririsan
-- ============================================================================
SELECT
    r.id,
    r.name,
    r.capacity
FROM rooms r
WHERE r.is_active = TRUE
  AND NOT EXISTS (
      SELECT 1
      FROM bookings b
      WHERE b.room_id = r.id
        AND b.status = 'CONFIRMED'
        AND b.start_time < '2026-12-01T12:00:00Z'::timestamptz
        AND b.end_time > '2026-12-01T10:00:00Z'::timestamptz
  )
ORDER BY r.id ASC;
