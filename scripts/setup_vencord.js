/**
 * setup_vencord.js
 * Downloads latest Vencord assets and prepares the Discord Electron Shim Loader.
 * Real Discord: _app.asar
 * Shim: app.asar (with index.js -> require("./patcher.js"))
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const VENCORD_REPO = 'Vendicated/Vencord';
const REQUIRED_FILES = ['patcher.js', 'preload.js', 'renderer.js', 'renderer.css'];

function getJson(url) {
  return new Promise((resolve, reject) => {
    https.get(url, {
      headers: {
        'User-Agent': 'DiscordPTB-Portable-Vencord-Setup',
        'Accept': 'application/vnd.github+json'
      }
    }, res => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return getJson(res.headers.location).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`Failed to fetch ${url}: HTTP ${res.statusCode}`));
      }
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (e) {
          reject(e);
        }
      });
    }).on('error', reject);
  });
}

function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    https.get(url, {
      headers: {
        'User-Agent': 'DiscordPTB-Portable-Vencord-Setup'
      }
    }, res => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return downloadFile(res.headers.location, destPath).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`Download failed for ${url}: HTTP ${res.statusCode}`));
      }
      const fileStream = fs.createWriteStream(destPath);
      res.pipe(fileStream);
      fileStream.on('finish', () => {
        fileStream.close();
        resolve(destPath);
      });
    }).on('error', reject);
  });
}

async function getLatestVencordRelease() {
  // Try /releases/latest first, fallback to /releases
  try {
    const latest = await getJson(`https://api.github.com/repos/${VENCORD_REPO}/releases/latest`);
    if (latest && latest.assets && latest.assets.length > 0) {
      return latest;
    }
  } catch (err) {
    console.warn('releases/latest error, trying /releases list:', err.message);
  }

  const releases = await getJson(`https://api.github.com/repos/${VENCORD_REPO}/releases`);
  if (!Array.isArray(releases) || releases.length === 0) {
    throw new Error('No Vencord releases found');
  }

  // Find first release with required assets
  for (const rel of releases) {
    if (rel.assets && rel.assets.some(a => a.name === 'patcher.js')) {
      return rel;
    }
  }

  return releases[0];
}

async function main() {
  const targetDir = process.argv[2] ? path.resolve(process.argv[2]) : path.join(__dirname, '..', 'res', 'vencord');
  fs.mkdirSync(targetDir, { recursive: true });

  console.log('Fetching latest Vencord release information...');
  const release = await getLatestVencordRelease();
  console.log(`Found release: ${release.tag_name} (${release.name || ''})`);

  const assetMap = new Map();
  for (const asset of release.assets || []) {
    assetMap.set(asset.name, asset.browser_download_url);
  }

  for (const file of REQUIRED_FILES) {
    const url = assetMap.get(file);
    if (!url) {
      // Fallback direct raw/download link
      const fallbackUrl = `https://github.com/${VENCORD_REPO}/releases/download/${release.tag_name}/${file}`;
      console.log(`Downloading ${file} from fallback ${fallbackUrl}...`);
      await downloadFile(fallbackUrl, path.join(targetDir, file));
    } else {
      console.log(`Downloading ${file} from ${url}...`);
      await downloadFile(url, path.join(targetDir, file));
    }
  }

  // Ensure shim files
  const packageJsonPath = path.join(targetDir, 'package.json');
  const indexJsPath = path.join(targetDir, 'index.js');
  fs.writeFileSync(packageJsonPath, JSON.stringify({ name: 'discord', main: 'index.js' }, null, 2) + '\n');
  fs.writeFileSync(indexJsPath, '// Vencord Shim Loader for Discord PTB Portable\nrequire("./patcher.js");\n');

  console.log(`Successfully prepared Vencord shim in: ${targetDir}`);

  // If --pack-asar flag was passed and asar is installed, build app.asar
  if (process.argv.includes('--pack-asar')) {
    const asarDest = path.join(targetDir, '..', 'app.asar');
    console.log(`Packaging shim into ${asarDest}...`);
    try {
      execSync(`asar pack "${targetDir}" "${asarDest}"`, { stdio: 'inherit' });
      console.log(`Successfully built ${asarDest}`);
    } catch (e) {
      console.warn('asar pack failed:', e.message);
    }
  }
}

if (require.main === module) {
  main().catch(err => {
    console.error('Error during Vencord setup:', err);
    process.exit(1);
  });
}

module.exports = {
  getLatestVencordRelease,
  REQUIRED_FILES,
  downloadFile
};
