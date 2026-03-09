create table if not exists users (
    id serial primary key,
    name varchar(255) not null,
    email varchar(255) not null unique,
    gender varchar(10),
    birth_date timestamp,
    created_at timestamp default current_timestamp
);

create table if not exists user_friends (
    user_id integer not null references users(id) on delete cascade,
    friend_id integer not null references users(id) on delete cascade,
    primary key (user_id, friend_id),
    check (user_id != friend_id)
);

create index if not exists idx_user_friends_friend_id on user_friends(friend_id);
create index if not exists idx_users_gender on users(gender);
create index if not exists idx_users_name on users(name);

insert into users (name, email, gender, birth_date) values 
('John Doe', 'john@example.com', 'male', '1990-01-15'),
('Jane Smith', 'jane@example.com', 'female', '1992-05-20'),
('Bob Johnson', 'bob@example.com', 'male', '1985-08-10'),
('Alice Williams', 'alice@example.com', 'female', '1995-03-25'),
('Charlie Brown', 'charlie@example.com', 'male', '1988-11-30'),
('Diana Prince', 'diana@example.com', 'female', '1991-07-12'),
('Edward Norton', 'edward@example.com', 'male', '1987-09-18'),
('Fiona Green', 'fiona@example.com', 'female', '1993-04-22'),
('George Miller', 'george@example.com', 'male', '1986-12-08'),
('Hannah Clark', 'hannah@example.com', 'female', '1994-06-14'),
('Isaac Newton', 'isaac@example.com', 'male', '1989-02-27'),
('Julia Roberts', 'julia@example.com', 'female', '1990-10-31'),
('Kevin Hart', 'kevin@example.com', 'male', '1988-07-06'),
('Laura Palmer', 'laura@example.com', 'female', '1996-01-19'),
('Michael Scott', 'michael@example.com', 'male', '1984-03-11'),
('Nicole Kidman', 'nicole@example.com', 'female', '1993-09-23'),
('Oliver Stone', 'oliver@example.com', 'male', '1987-05-16'),
('Patricia Hill', 'patricia@example.com', 'female', '1991-12-05'),
('Quinn Adams', 'quinn@example.com', 'male', '1989-08-29'),
('Rachel Green', 'rachel@example.com', 'female', '1992-11-02'),
('Samuel Jackson', 'samuel@example.com', 'male', '1986-04-13'),
('Tina Turner', 'tina@example.com', 'female', '1995-02-26');

-- Friend Connections
-- John (1) имеет друзей: Jane (2), Alice (4), Diana (6) - 3 друга
insert into user_friends (user_id, friend_id) values 
(1, 2), (2, 1),  -- John ↔ Jane
(1, 4), (4, 1),  -- John ↔ Alice
(1, 6), (6, 1),  -- John ↔ Diana
(1, 8), (8, 1),  -- John ↔ Fiona

-- Bob (3) имеет друзей: Alice (4), Diana (6), Hannah (10) - 3 друга
(3, 4), (4, 3),  -- Bob ↔ Alice
(3, 6), (6, 3),  -- Bob ↔ Diana
(3, 10), (10, 3),  -- Bob ↔ Hannah
(3, 12), (12, 3),  -- Bob ↔ Julia

-- Charlie (5) имеет друзей: Diana (6), Fiona (8), Hannah (10)
(5, 6), (6, 5),  -- Charlie ↔ Diana
(5, 8), (8, 5),  -- Charlie ↔ Fiona
(5, 10), (10, 5),  -- Charlie ↔ Hannah
(5, 14), (14, 5),  -- Charlie ↔ Laura

-- George (9) имеет друзей: Hannah (10), Julia (12), Kevin (13)
(9, 10), (10, 9),  -- George ↔ Hannah
(9, 12), (12, 9),  -- George ↔ Julia
(9, 13), (13, 9),  -- George ↔ Kevin
(9, 15), (15, 9),  -- George ↔ Michael

-- Additional connections
(2, 4), (4, 2),  -- Jane ↔ Alice
(2, 6), (6, 2),  -- Jane ↔ Diana
(4, 6), (6, 4),  -- Alice ↔ Diana
(4, 8), (8, 4),  -- Alice ↔ Fiona
(6, 8), (8, 6),  -- Diana ↔ Fiona
(7, 9), (9, 7),  -- Edward ↔ George
(7, 11), (11, 7),  -- Edward ↔ Isaac
(11, 13), (13, 11),  -- Isaac ↔ Kevin
(13, 15), (15, 13),  -- Kevin ↔ Michael
(10, 12), (12, 10),  -- Hannah ↔ Julia
(14, 16), (16, 14),  -- Laura ↔ Nicole
(16, 18), (18, 16),  -- Nicole ↔ Patricia
(17, 19), (19, 17);  -- Oliver ↔ Quinn