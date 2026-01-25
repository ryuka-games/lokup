package report

// htmlTemplate は HTML レポートのテンプレート。
const htmlTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Lokup レポート - {{.Repository}}</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f5f5f5;
            color: #333;
            line-height: 1.6;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 20px;
            text-align: center;
        }
        header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
        }
        header .subtitle {
            opacity: 0.9;
            font-size: 1.1rem;
        }
        .meta {
            display: flex;
            justify-content: center;
            gap: 30px;
            margin-top: 20px;
            font-size: 0.95rem;
        }
        .section {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin: 20px 0;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
        }
        .section h2 {
            font-size: 1.5rem;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #eee;
        }
        .scores {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 20px;
        }
        .score-card {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 25px;
            text-align: center;
        }
        .score-card h3 {
            font-size: 1rem;
            color: #666;
            margin-bottom: 15px;
        }
        .score-value {
            font-size: 3rem;
            font-weight: bold;
        }
        .score-value.grade-a { color: #22c55e; }
        .score-value.grade-b { color: #84cc16; }
        .score-value.grade-c { color: #eab308; }
        .score-value.grade-d { color: #ef4444; }
        .score-grade {
            font-size: 1.2rem;
            margin-top: 5px;
            color: #666;
        }
        .score-desc {
            font-size: 0.85rem;
            color: #888;
            margin-top: 15px;
            line-height: 1.5;
        }
        .section-desc {
            color: #666;
            margin-bottom: 20px;
            font-size: 0.95rem;
        }
        .criteria {
            font-size: 0.8rem;
            color: #999;
        }
        .score-breakdown {
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid #eee;
        }
        .score-breakdown table {
            width: 100%;
            font-size: 0.85rem;
        }
        .score-breakdown td {
            padding: 4px 0;
        }
        .score-breakdown .points {
            text-align: right;
            font-weight: bold;
        }
        .score-breakdown .positive .points {
            color: #22c55e;
        }
        .score-breakdown .negative .points {
            color: #ef4444;
        }
        .score-breakdown .total {
            border-top: 1px solid #ddd;
            font-weight: bold;
        }
        .score-breakdown .total td {
            padding-top: 8px;
        }
        .score-breakdown .detail {
            color: #999;
            font-size: 0.8rem;
        }
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
        }
        .metric-card {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            text-align: center;
        }
        .metric-card .value {
            font-size: 2rem;
            font-weight: bold;
            color: #667eea;
        }
        .metric-card .label {
            font-size: 0.9rem;
            color: #666;
            margin-top: 5px;
        }
        .metric-card .metric-desc {
            font-size: 0.8rem;
            color: #999;
            margin-top: 10px;
            line-height: 1.4;
        }
        .metric-card.warning {
            border: 2px solid #eab308;
            background: #fefce8;
        }
        .metric-card.warning .value {
            color: #ca8a04;
        }
        .risks-list {
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        .risk-item {
            display: flex;
            align-items: flex-start;
            gap: 15px;
            padding: 15px;
            border-radius: 8px;
            background: #f8f9fa;
        }
        .risk-item.high { border-left: 4px solid #ef4444; }
        .risk-item.medium { border-left: 4px solid #eab308; }
        .risk-item.low { border-left: 4px solid #22c55e; }
        .risk-icon {
            font-size: 1.5rem;
        }
        .risk-content h4 {
            font-size: 1rem;
            margin-bottom: 5px;
        }
        .risk-content p {
            font-size: 0.9rem;
            color: #666;
        }
        .risk-content .risk-action {
            margin-top: 10px;
            padding: 10px;
            background: #f0f9ff;
            border-radius: 6px;
            color: #0369a1;
            font-size: 0.85rem;
        }
        .no-risks {
            text-align: center;
            padding: 40px;
            color: #22c55e;
            font-size: 1.1rem;
        }
        .chart-container {
            position: relative;
            height: 300px;
            margin-top: 20px;
        }
        footer {
            text-align: center;
            padding: 30px;
            color: #999;
            font-size: 0.85rem;
        }
        @media (max-width: 768px) {
            header h1 { font-size: 1.8rem; }
            .meta { flex-direction: column; gap: 10px; }
            .score-value { font-size: 2.5rem; }
        }
    </style>
</head>
<body>
    <header>
        <h1>{{.Repository}}</h1>
        <p class="subtitle">GitHub リポジトリ健康診断レポート</p>
        <div class="meta">
            <span>分析期間: {{.PeriodFrom}} ~ {{.PeriodTo}} ({{.PeriodDays}}日間)</span>
            <span>生成日時: {{.GeneratedAt}}</span>
        </div>
    </header>

    <div class="container">
        <!-- Executive Summary -->
        <section class="section">
            <h2>概要</h2>
            <p class="section-desc">このリポジトリの健全性を100点満点で評価した結果です。</p>
            <div class="scores">
                <div class="score-card">
                    <h3>開発効率スコア</h3>
                    <div class="score-value grade-{{.EfficiencyGrade | lower}}">{{.EfficiencyScore}}</div>
                    <div class="score-grade">グレード {{.EfficiencyGrade}}</div>
                    <div class="score-breakdown">
                        <table>
                            {{range .EfficiencyBreakdown}}
                            <tr class="{{if gt .Points 0}}positive{{else if lt .Points 0}}negative{{end}}">
                                <td>{{.Label}}{{if .Detail}} <span class="detail">({{.Detail}})</span>{{end}}</td>
                                <td class="points">{{if gt .Points 0}}+{{end}}{{.Points}}</td>
                            </tr>
                            {{end}}
                            <tr class="total">
                                <td>合計</td>
                                <td class="points">{{.EfficiencyScore}}</td>
                            </tr>
                        </table>
                    </div>
                </div>
                <div class="score-card">
                    <h3>コード健全性スコア</h3>
                    <div class="score-value grade-{{.HealthGrade | lower}}">{{.HealthScore}}</div>
                    <div class="score-grade">グレード {{.HealthGrade}}</div>
                    <div class="score-breakdown">
                        <table>
                            {{range .HealthBreakdown}}
                            <tr class="{{if gt .Points 0}}positive{{else if lt .Points 0}}negative{{end}}">
                                <td>{{.Label}}{{if .Detail}} <span class="detail">({{.Detail}})</span>{{end}}</td>
                                <td class="points">{{if gt .Points 0}}+{{end}}{{.Points}}</td>
                            </tr>
                            {{end}}
                            <tr class="total">
                                <td>合計</td>
                                <td class="points">{{.HealthScore}}</td>
                            </tr>
                        </table>
                    </div>
                </div>
            </div>
        </section>

        <!-- Metrics -->
        <section class="section">
            <h2>メトリクス</h2>
            <p class="section-desc">分析期間中の主要な指標です。</p>
            <div class="metrics">
                <div class="metric-card">
                    <div class="value">{{.TotalCommits}}</div>
                    <div class="label">総コミット数</div>
                    <div class="metric-desc">期間中にマージされた変更の総数</div>
                </div>
                <div class="metric-card">
                    <div class="value">{{printf "%.2f" .FeatureAddition}}</div>
                    <div class="label">コミット/日</div>
                    <div class="metric-desc">1日あたりの平均コミット数<br>開発の活発さを示す</div>
                </div>
                <div class="metric-card">
                    <div class="value">{{.Contributors}}</div>
                    <div class="label">コントリビューター数</div>
                    <div class="metric-desc">コードに貢献した人数<br>多いほど属人化リスクが低い</div>
                </div>
                <div class="metric-card {{if ge .LateNightRate 30.0}}warning{{end}}">
                    <div class="value">{{printf "%.1f" .LateNightRate}}%</div>
                    <div class="label">深夜コミット率</div>
                    <div class="metric-desc">22:00〜翌5:00のコミット割合<br><span class="criteria">基準: 30%以下が健全</span></div>
                </div>
                <div class="metric-card {{if ge .AvgLeadTime 7.0}}warning{{end}}">
                    <div class="value">{{printf "%.1f" .AvgLeadTime}}日</div>
                    <div class="label">PRリードタイム</div>
                    <div class="metric-desc">PR作成からマージまでの平均日数<br><span class="criteria">基準: 7日以下が健全</span></div>
                </div>
            </div>
        </section>

        <!-- Risks -->
        <section class="section">
            <h2>検出されたリスク</h2>
            <p class="section-desc">自動検出された問題点と改善提案です。</p>
            {{if .HasRisks}}
            <div class="risks-list">
                {{range .Risks}}
                <div class="risk-item {{.Severity}}">
                    <span class="risk-icon">{{.SeverityIcon}}</span>
                    <div class="risk-content">
                        <h4>{{.Type}}</h4>
                        <p class="risk-description">{{.Description}}</p>
                        {{if .Target}}<p class="risk-target"><strong>対象:</strong> {{.Target}}</p>{{end}}
                        <p class="risk-action">{{.Action}}</p>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="no-risks">
                重大なリスクは検出されませんでした
            </div>
            {{end}}
        </section>

        <!-- Score Chart -->
        <section class="section">
            <h2>スコア概要</h2>
            <div class="chart-container">
                <canvas id="scoreChart"></canvas>
            </div>
        </section>

        <!-- Daily Commits Chart -->
        <section class="section">
            <h2>日別コミット推移</h2>
            <p class="section-desc">分析期間中の日々のコミット数の推移です。週末や締切前の傾向が見えます。</p>
            <div class="chart-container">
                <canvas id="dailyCommitsChart"></canvas>
            </div>
        </section>
    </div>

    <footer>
        <p>Lokup - GitHub リポジトリ健康診断ツール</p>
    </footer>

    <script>
        // Score Chart
        new Chart(document.getElementById('scoreChart'), {
            type: 'bar',
            data: {
                labels: ['開発効率', 'コード健全性'],
                datasets: [{
                    label: 'スコア',
                    data: [{{.EfficiencyScore}}, {{.HealthScore}}],
                    backgroundColor: [
                        'rgba(102, 126, 234, 0.8)',
                        'rgba(118, 75, 162, 0.8)'
                    ],
                    borderColor: [
                        'rgb(102, 126, 234)',
                        'rgb(118, 75, 162)'
                    ],
                    borderWidth: 1,
                    borderRadius: 8
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        ticks: {
                            stepSize: 20
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });

        // Daily Commits Chart
        new Chart(document.getElementById('dailyCommitsChart'), {
            type: 'line',
            data: {
                labels: [{{range $i, $label := .CommitDayLabels}}{{if $i}}, {{end}}'{{$label}}'{{end}}],
                datasets: [{
                    label: 'コミット数',
                    data: [{{range $i, $count := .CommitsByDay}}{{if $i}}, {{end}}{{$count}}{{end}}],
                    borderColor: 'rgb(102, 126, 234)',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    fill: true,
                    tension: 0.3,
                    pointRadius: 4,
                    pointHoverRadius: 6
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            stepSize: 1
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    </script>
</body>
</html>`
