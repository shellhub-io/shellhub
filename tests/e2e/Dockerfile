FROM mcr.microsoft.com/playwright:v1.57.0-jammy

WORKDIR /app

# Copy package files
COPY package.json ./

# Install dependencies
RUN npm install

# Copy test files and config
COPY . .

# Default command runs tests
CMD ["npm", "test"]
