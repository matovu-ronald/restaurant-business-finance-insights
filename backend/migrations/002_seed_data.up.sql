-- 002_seed_data.up.sql
-- Seed default service channels and dayparts

-- Service channels
INSERT INTO service_channels (id, code, display_name) VALUES
    (uuid_generate_v4(), 'dine-in', 'Dine In'),
    (uuid_generate_v4(), 'takeaway', 'Takeaway'),
    (uuid_generate_v4(), 'pickup', 'Online Pickup'),
    (uuid_generate_v4(), 'catering', 'Catering');

-- Dayparts (Brisbane timezone defaults)
INSERT INTO dayparts (id, code, display_name, start_time, end_time) VALUES
    (uuid_generate_v4(), 'breakfast', 'Breakfast', '07:00:00', '11:00:00'),
    (uuid_generate_v4(), 'lunch', 'Lunch', '11:00:00', '15:00:00'),
    (uuid_generate_v4(), 'dinner', 'Dinner', '15:00:00', '20:00:00');

-- Default location (The Lakehouse)
INSERT INTO locations (id, name, timezone, seating_capacity_indoor, seating_capacity_patio) VALUES
    (uuid_generate_v4(), 'The Lakehouse Restaurant', 'Australia/Brisbane', 70, 30);
