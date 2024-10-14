-- Create categories table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES categories(id),
    level INTEGER NOT NULL
);

-- Create pantries table
CREATE TABLE pantries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    sub_categories TEXT[],
    description TEXT,
    image_urls TEXT[]
);

-- Create products table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL,
    pantry_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INTEGER NOT NULL,
    image_urls TEXT[]
);

CREATE TABLE recipes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    difficulty VARCHAR(50),
    prep_time INTEGER,
    cook_time INTEGER,
    servings INTEGER,
    category VARCHAR(255),
    sub_categories TEXT[],
    image_urls TEXT[],
    is_public BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE recipe_ingredient (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL,
    pantry_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL
);

CREATE TABLE recipe_tool (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL,
    pantry_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL
);

CREATE TABLE steps (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL,
    step_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT
);