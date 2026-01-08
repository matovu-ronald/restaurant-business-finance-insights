-- 002_seed_data.down.sql
DELETE FROM locations WHERE name = 'The Lakehouse Restaurant';
DELETE FROM dayparts WHERE code IN ('breakfast', 'lunch', 'dinner');
DELETE FROM service_channels WHERE code IN ('dine-in', 'takeaway', 'pickup', 'catering');
