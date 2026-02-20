create table if not exists users (
    id serial primary key,
    name varchar(255) not null,
    email varchar(255) not null unique,
    age int,
    city varchar(255),
    created_at timestamp default current_timestamp
);

insert into users (name, email, age, city) values 
('John Doe', 'john@example.com', 30, 'New York'),
('Jane Smith', 'jane@example.com', 28, 'Los Angeles'),
('Bob Johnson', 'bob@example.com', 35, 'Chicago');