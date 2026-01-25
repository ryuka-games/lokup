/**
 * GitHub API 検証スクリプト
 * 診断に必要なデータが取得できるか確認する
 */

const REPO = 'facebook/react';
const BASE_URL = 'https://api.github.com';

async function fetchGitHub(endpoint) {
  const res = await fetch(`${BASE_URL}${endpoint}`, {
    headers: {
      'Accept': 'application/vnd.github.v3+json',
      'User-Agent': 'lokup-api-test'
    }
  });

  // レート制限の確認
  const remaining = res.headers.get('x-ratelimit-remaining');
  const limit = res.headers.get('x-ratelimit-limit');
  console.log(`  [Rate Limit: ${remaining}/${limit}]`);

  if (!res.ok) {
    throw new Error(`API Error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

async function testCommits() {
  console.log('\n📌 1. コミット履歴の取得');
  console.log('─'.repeat(40));

  const commits = await fetchGitHub(`/repos/${REPO}/commits?per_page=5`);

  console.log(`  取得件数: ${commits.length}`);
  console.log('  サンプル:');

  const c = commits[0];
  console.log(`    - SHA: ${c.sha.slice(0, 7)}`);
  console.log(`    - Author: ${c.commit.author.name}`);
  console.log(`    - Date: ${c.commit.author.date}`);
  console.log(`    - Message: ${c.commit.message.split('\n')[0].slice(0, 50)}...`);

  return { success: true, note: '日時、author、メッセージ取得可能' };
}

async function testCommitDetail() {
  console.log('\n📌 2. コミット詳細（変更ファイル）');
  console.log('─'.repeat(40));

  // まず最新コミットのSHAを取得
  const commits = await fetchGitHub(`/repos/${REPO}/commits?per_page=1`);
  const sha = commits[0].sha;

  const detail = await fetchGitHub(`/repos/${REPO}/commits/${sha}`);

  console.log(`  変更ファイル数: ${detail.files?.length || 0}`);
  if (detail.files && detail.files.length > 0) {
    const f = detail.files[0];
    console.log('  サンプル:');
    console.log(`    - File: ${f.filename}`);
    console.log(`    - Additions: ${f.additions}, Deletions: ${f.deletions}`);
    console.log(`    - Status: ${f.status}`);
  }

  return { success: true, note: '変更ファイル、追加/削除行数取得可能' };
}

async function testPullRequests() {
  console.log('\n📌 3. PR情報');
  console.log('─'.repeat(40));

  const prs = await fetchGitHub(`/repos/${REPO}/pulls?state=closed&per_page=5`);

  // マージされたPRを探す
  const mergedPR = prs.find(pr => pr.merged_at);

  if (mergedPR) {
    console.log('  マージ済みPRサンプル:');
    console.log(`    - Title: ${mergedPR.title.slice(0, 50)}...`);
    console.log(`    - Created: ${mergedPR.created_at}`);
    console.log(`    - Merged: ${mergedPR.merged_at}`);

    const created = new Date(mergedPR.created_at);
    const merged = new Date(mergedPR.merged_at);
    const days = ((merged - created) / (1000 * 60 * 60 * 24)).toFixed(1);
    console.log(`    - リードタイム: ${days}日`);
  }

  return { success: true, note: 'PR作成日→マージ日取得可能、リードタイム計算可能' };
}

async function testFileContent() {
  console.log('\n📌 4. ファイル内容（行数確認用）');
  console.log('─'.repeat(40));

  const content = await fetchGitHub(`/repos/${REPO}/contents/package.json`);

  const decoded = Buffer.from(content.content, 'base64').toString('utf-8');
  const lines = decoded.split('\n').length;

  console.log(`  ファイル: package.json`);
  console.log(`  サイズ: ${content.size} bytes`);
  console.log(`  行数: ${lines}`);

  return { success: true, note: 'ファイル内容取得可能、行数カウント可能' };
}

async function testDependencies() {
  console.log('\n📌 5. 依存パッケージ');
  console.log('─'.repeat(40));

  const content = await fetchGitHub(`/repos/${REPO}/contents/package.json`);
  const decoded = Buffer.from(content.content, 'base64').toString('utf-8');
  const pkg = JSON.parse(decoded);

  console.log('  devDependencies (一部):');
  const deps = Object.entries(pkg.devDependencies || {}).slice(0, 5);
  deps.forEach(([name, version]) => {
    console.log(`    - ${name}: ${version}`);
  });

  return { success: true, note: 'package.json から依存バージョン取得可能' };
}

async function testContributors() {
  console.log('\n📌 6. コントリビューター（属人化確認用）');
  console.log('─'.repeat(40));

  const contributors = await fetchGitHub(`/repos/${REPO}/contributors?per_page=10`);

  console.log('  Top 5 Contributors:');
  contributors.slice(0, 5).forEach((c, i) => {
    console.log(`    ${i + 1}. ${c.login}: ${c.contributions} commits`);
  });

  return { success: true, note: 'コントリビューター別コミット数取得可能' };
}

// メイン実行
async function main() {
  console.log('='.repeat(50));
  console.log('GitHub API 検証: ' + REPO);
  console.log('='.repeat(50));

  const results = [];

  try {
    results.push({ name: 'コミット履歴', ...await testCommits() });
    results.push({ name: 'コミット詳細', ...await testCommitDetail() });
    results.push({ name: 'PR情報', ...await testPullRequests() });
    results.push({ name: 'ファイル内容', ...await testFileContent() });
    results.push({ name: '依存パッケージ', ...await testDependencies() });
    results.push({ name: 'コントリビューター', ...await testContributors() });
  } catch (e) {
    console.error('\n❌ エラー:', e.message);
  }

  console.log('\n' + '='.repeat(50));
  console.log('📊 検証結果サマリー');
  console.log('='.repeat(50));

  results.forEach(r => {
    const icon = r.success ? '✅' : '❌';
    console.log(`${icon} ${r.name}: ${r.note}`);
  });
}

main();
