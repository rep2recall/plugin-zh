import { execSync } from 'child_process'
import fs from 'fs'
import https from 'https'
import os from 'os'
import path from 'path'

import sqlite3 from 'better-sqlite3'

import { db } from './shared'

export async function populate(
  db: sqlite3.Database,
  dir = fs.mkdtempSync(path.join(os.tmpdir(), 'cedict'))
) {
  process.chdir(dir)

  db.exec(/* sql */ `
  CREATE TABLE IF NOT EXISTS "cedict" (
    "simplified"    TEXT NOT NULL,
    "traditional"   TEXT CHECK ("simplified" != "traditional"),
    "reading"       TEXT,
    "english"       JSON
  );

  CREATE UNIQUE INDEX IF NOT EXISTS idx_u_cedict ON "cedict" ("simplified", "traditional", "reading");
  `)

  try {
    console.log('Downloading the latest CEDICT.')

    const zipName = './cedict_1_0_ts_utf-8_mdbg.txt.gz'
    const urlString =
      'https://www.mdbg.net/chinese/export/cedict/cedict_1_0_ts_utf-8_mdbg.txt.gz'
    if (fs.existsSync(zipName)) {
      fs.unlinkSync(zipName)
    }
    const f = fs.createWriteStream(zipName)
    https.get(urlString, (res) => {
      res.pipe(f)
    })

    await new Promise((resolve, reject) => {
      f.once('error', reject).once('finish', resolve)
    })

    execSync(`gzip -d ${zipName}`)

    const f2 = fs.createReadStream('./cedict_1_0_ts_utf-8_mdbg.txt')
    db.exec('BEGIN')
    const stmt = db.prepare(/* sql */ `
    INSERT INTO "cedict" ("simplified", "traditional", "reading", "english")
    VALUES (@simplified, @traditional, @reading, @english)
    ON CONFLICT DO NOTHING
    `)

    let line = ''
    f2.on('data', (d) => {
      const lines = (line + d.toString()).split('\n')
      line = lines.pop() || ''

      lines.map((ln) => {
        const m = /^(\p{sc=Han}+) (\p{sc=Han}+) \[([^\]]+)\] \/(.+)\/$/u.exec(
          ln.trim()
        )

        if (m) {
          stmt.run({
            simplified: m[2],
            traditional: m[2] === m[1] ? null : m[1],
            reading: m[3],
            english: JSON.stringify(m[4]!.split('/'))
          })
        }
      })
    })

    await new Promise<void>((resolve, reject) => {
      f2.once('error', reject).once('end', () => {
        const m = /^(\p{sc=Han}+) (\p{sc=Han}+) \[([^\]]+)\] \/(.+)\/$/u.exec(
          line.trim()
        )

        if (m) {
          stmt.run({
            simplified: m[2],
            traditional: m[2] === m[1] ? null : m[1],
            reading: m[3],
            english: JSON.stringify(m[4]!.split('/'))
          })
        }

        resolve()
      })
    })

    db.exec('COMMIT')
  } catch (e) {
    console.error(e)
  }

  db.close()
}

if (require.main === module) {
  populate(db)
}
