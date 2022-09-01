process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = "0";

import fs from 'fs'
import path     from 'path'
import { QBittorrent } from '@ctrl/qbittorrent';

const DOWNLOAD_DIRS = process.env.DOWNLOAD_DIRS.split(',');

const client = new QBittorrent({
   baseUrl: process.env.SERVER_URL,
   username: process.env.SERVER_USER,
   password: process.env.SERVER_PASS,
 });

const torrents = await client.listTorrents()

if (!torrents?.length) {
  console.log('No torrents found')
  process.exit(0)
}

for (const torrent of torrents) {
  if (!torrent.completed) {
    console.log(`Skipping because it's not complete: ${torrent.name} `)
    continue;
  }
  console.log(`Verifying if all are files present: ${torrent.name}`)
  const files = await client.torrentFiles(torrent.hash)
  let missing = false;
  for (const file of files) {
    if (file.priority === 0) {
      // Skip files that are not downloaded
      continue;
    }
    let found = false;
    DOWNLOAD_DIRS.some(async dir => {
      for (const dir of DOWNLOAD_DIRS) {
        found = found || fs.existsSync(path.join(dir, file.name))
      }
      if (!found) {
        console.log(`File ${file.name} is missing for ${torrent.name}, removing the torrent from qBT`)
        missing = true;
        await client.removeTorrent(torrent.hash, true)
      }
      return !found;
    })
    if (!found) {
      break;
    }
  }
  if (!missing) {
    console.log(`All files are present for ${torrent.name}`)
  }
}