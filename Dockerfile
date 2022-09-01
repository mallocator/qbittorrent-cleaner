FROM node:18-alpine

WORKDIR /app

COPY package.json index.js ./

RUN npm install --omit=dev

#Specify one ore more directories to watch for changes
ENV DOWNLOAD_DIRS=/downloads
# The URL of your qbittorent server running the WebUI
ENV SERVER_URL=https://10.0.0.1:8080
# The username configured in the WebUI for qbittorent
ENV SERVER_USER=admin
# The password configured in the WebUI for qbittorent
ENV SERVER_PASS=adminadmin

# The image will terminate at the end. This image is expected to be run mutiple times.
CMD ["npm", "start"]
