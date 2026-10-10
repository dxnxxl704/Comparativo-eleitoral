const dashboardMetrics = document.getElementById('dashboardMetrics');
const candidateList = document.getElementById('candidateList');
const comparisonResults = document.getElementById('comparisonResults');
const reportList = document.getElementById('reportList');

const api = {
  async listCandidates(params = '') {
    const response = await fetch(`/api/candidates${params}`);
    return response.json();
  },
  async listDashboard() {
    const response = await fetch('/api/dashboard');
    return response.json();
  },
  async compareCandidates(ids) {
    const response = await fetch(`/api/candidates/compare?ids=${encodeURIComponent(ids)}`);
    return response.json();
  },
  async createCandidate(payload) {
    const response = await fetch('/api/candidates', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return response.json();
  },
  async createEvaluation(payload) {
    const response = await fetch('/api/evaluations', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return response.json();
  },
  async createSource(payload) {
    const response = await fetch('/api/sources', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return response.json();
  },
  async getReports() {
    const response = await fetch('/api/reports');
    return response.json();
  },
};

function renderMetricCard(label, value, accentClass) {
  const card = document.createElement('div');
  card.className = `metric-card rounded-2xl p-5 shadow-sm ring-1 ${accentClass}`;
  card.innerHTML = `
    <p class="text-sm uppercase tracking-[0.16em] text-slate-500">${label}</p>
    <div class="mt-3 text-3xl font-bold text-slate-900">${value}</div>
  `;
  return card;
}

function renderDashboard(dashboard) {
  dashboardMetrics.innerHTML = '';
  dashboardMetrics.appendChild(renderMetricCard('Candidatos', dashboard.total_candidates || 0, 'ring-sky-200'));
  dashboardMetrics.appendChild(renderMetricCard('Avaliações', dashboard.total_evaluations || 0, 'ring-violet-200'));
  dashboardMetrics.appendChild(renderMetricCard('Fontes', dashboard.total_sources || 0, 'ring-amber-200'));
  dashboardMetrics.appendChild(renderMetricCard('Coerência média', `${Number(dashboard.average_coherence || 0).toFixed(1)}%`, 'ring-emerald-200'));
}

function renderCandidates(candidates) {
  candidateList.innerHTML = '';
  if (!Array.isArray(candidates) || candidates.length === 0) {
    candidateList.innerHTML = '<div class="rounded-xl border border-dashed border-slate-300 bg-slate-50 p-5 text-sm text-slate-500">Nenhum candidato encontrado.</div>';
    return;
  }

  candidates.forEach((candidate) => {
    const card = document.createElement('article');
    card.className = 'candidate-card rounded-xl border border-slate-200 bg-slate-50 p-4';
    card.innerHTML = `
      <div class="flex items-start justify-between gap-3">
        <div>
          <div class="text-lg font-semibold text-slate-900">${candidate.name}</div>
          <div class="text-sm text-slate-500">${candidate.party} • ${candidate.office} • ${candidate.state}</div>
        </div>
        <span class="score-badge rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold text-emerald-700">ID #${candidate.id}</span>
      </div>
      <p class="mt-3 text-sm text-slate-600">${candidate.biography || 'Sem biografia cadastrada.'}</p>
      <div class="mt-4 flex flex-wrap gap-2">
        <button data-report-id="${candidate.id}" class="report-button rounded-lg border border-sky-200 bg-sky-50 px-3 py-1.5 text-xs font-semibold text-sky-700 hover:bg-sky-100">Relatório</button>
        <button data-compare-id="${candidate.id}" class="compare-button rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-semibold text-emerald-700 hover:bg-emerald-100">Comparar</button>
      </div>
    `;
    candidateList.appendChild(card);
  });

  document.querySelectorAll('.report-button').forEach((button) => {
    button.addEventListener('click', async () => {
      const report = await fetch(`/api/reports/${button.dataset.reportId}`).then((res) => res.json());
      renderSingleReport(report);
    });
  });

  document.querySelectorAll('.compare-button').forEach((button) => {
    button.addEventListener('click', () => {
      const current = document.getElementById('compareIds').value.trim();
      const next = current ? `${current},${button.dataset.compareId}` : `${button.dataset.compareId}`;
      document.getElementById('compareIds').value = next;
      triggerComparison();
    });
  });
}

function renderSingleReport(report) {
  const container = document.createElement('div');
  container.className = 'report-card rounded-2xl p-4';
  const score = Number(report.coherence_score || 0).toFixed(1);
  const criteria = Object.entries(report.criteria || {}).map(([key, value]) => `${key}: ${value ? 'Sim' : 'Não'}`).join(' • ');
  const sources = (report.sources || []).map((source) => `<li><a href="${source.url}" target="_blank" class="text-sky-700 underline">${source.title}</a></li>`).join('') || '<li>Sem fontes vinculadas.</li>';
  container.innerHTML = `
    <div class="flex items-center justify-between gap-3">
      <div>
        <h4 class="text-lg font-semibold text-slate-900">${report.candidate.name}</h4>
        <p class="text-sm text-slate-500">${report.candidate.party} • ${report.candidate.office}</p>
      </div>
      <span class="score-badge rounded-full bg-emerald-100 px-3 py-1 text-sm font-bold text-emerald-700">${score}%</span>
    </div>
    <p class="mt-3 text-sm text-slate-600">${criteria}</p>
    <div class="mt-3">
      <p class="text-xs uppercase tracking-[0.16em] text-slate-500">Fontes</p>
      <ul class="mt-2 list-disc pl-5 text-sm text-slate-700">${sources}</ul>
    </div>
  `;
  reportList.innerHTML = '';
  reportList.appendChild(container);
}

function renderReports(reportsData) {
  const reports = Array.isArray(reportsData?.reports) ? reportsData.reports : [];
  reportList.innerHTML = '';

  if (reports.length === 0) {
    reportList.innerHTML = '<div class="rounded-xl border border-dashed border-slate-300 bg-slate-50 p-5 text-sm text-slate-500">Nenhum relatório disponível.</div>';
    return;
  }

  reports.forEach((report) => {
    const card = document.createElement('article');
    card.className = 'report-card rounded-2xl p-4';
    const score = Number(report.coherence_score || 0).toFixed(1);
    card.innerHTML = `
      <div class="flex items-center justify-between gap-3">
        <div>
          <h4 class="text-lg font-semibold text-slate-900">${report.candidate.name}</h4>
          <p class="text-sm text-slate-500">${report.candidate.party} • ${report.candidate.office}</p>
        </div>
        <span class="score-badge rounded-full bg-emerald-100 px-3 py-1 text-sm font-bold text-emerald-700">${score}%</span>
      </div>
      <p class="mt-3 text-sm text-slate-600">${report.evaluation ? report.evaluation.observacoes || 'Sem observações registradas.' : 'Sem avaliação registrada.'}</p>
      <div class="mt-3 flex flex-wrap gap-2 text-xs text-slate-600">
        ${Object.entries(report.criteria || {}).map(([key, value]) => '<span class="rounded-full bg-slate-100 px-2 py-1">' + key + ': ' + (value ? 'Sim' : 'Não') + '</span>').join('')}
      </div>
    `;
    reportList.appendChild(card);
  });
}

function renderComparison(results) {
  comparisonResults.innerHTML = '';
  if (!Array.isArray(results?.candidates) || results.candidates.length === 0) {
    comparisonResults.innerHTML = '<div class="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-3 text-sm text-slate-500">Nenhum dado de comparação.</div>';
    return;
  }

  results.candidates.forEach((slot) => {
    const item = document.createElement('div');
    item.className = 'rounded-lg border border-slate-200 bg-slate-50 p-3';
    item.innerHTML = `
      <div class="flex items-center justify-between gap-3">
        <div>
          <div class="font-semibold text-slate-800">${slot.candidate.name}</div>
          <div class="text-xs text-slate-500">${slot.candidate.office} • ${slot.candidate.state}</div>
        </div>
        <span class="rounded-full bg-emerald-100 px-2 py-1 text-xs font-bold text-emerald-700">${Number(slot.score || 0).toFixed(1)}%</span>
      </div>
    `;
    comparisonResults.appendChild(item);
  });
}

async function loadCandidates() {
  const query = document.getElementById('candidateSearch').value.trim();
  const params = new URLSearchParams();
  if (query) params.set('q', query);
  const filterFields = {
    name: document.getElementById('filterName').value.trim(),
    party: document.getElementById('filterParty').value.trim(),
    office: document.getElementById('filterOffice').value.trim(),
    state: document.getElementById('filterState').value.trim(),
  };
  Object.entries(filterFields).forEach(([key, value]) => {
    if (value) params.set(key, value);
  });

  const queryString = params.toString();
  const candidates = await api.listCandidates(queryString ? `?${queryString}` : '');
  renderCandidates(candidates);
}

async function loadDashboard() {
  const dashboard = await api.listDashboard();
  renderDashboard(dashboard);
}

async function loadReports() {
  const reports = await api.getReports();
  renderReports(reports);
}

async function triggerComparison() {
  const ids = document.getElementById('compareIds').value.trim();
  if (!ids) {
    comparisonResults.innerHTML = '<div class="rounded-lg border border-dashed border-slate-300 bg-slate-50 p-3 text-sm text-slate-500">Digite IDs para comparar.</div>';
    return;
  }
  const result = await api.compareCandidates(ids);
  renderComparison(result);
}

async function handleCandidateSubmit(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const payload = Object.fromEntries(new FormData(form).entries());
  await api.createCandidate(payload);
  form.reset();
  await loadCandidates();
  await loadDashboard();
  await loadReports();
}

async function handleEvaluationSubmit(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const payload = Object.fromEntries(new FormData(form).entries());
  payload.candidate_id = Number(payload.candidate_id);
  payload.nao_neoliberal = Boolean(payload.nao_neoliberal);
  payload.contra_aborto = Boolean(payload.contra_aborto);
  payload.respeita_cristianismo = Boolean(payload.respeita_cristianismo);
  await api.createEvaluation(payload);
  form.reset();
  await loadDashboard();
  await loadReports();
}

async function handleSourceSubmit(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const payload = Object.fromEntries(new FormData(form).entries());
  payload.candidate_id = Number(payload.candidate_id);
  await api.createSource(payload);
  form.reset();
  await loadDashboard();
  await loadReports();
}

document.getElementById('searchButton').addEventListener('click', loadCandidates);
document.getElementById('candidateSearch').addEventListener('input', loadCandidates);
document.getElementById('filterName').addEventListener('input', loadCandidates);
document.getElementById('filterParty').addEventListener('input', loadCandidates);
document.getElementById('filterOffice').addEventListener('input', loadCandidates);
document.getElementById('filterState').addEventListener('input', loadCandidates);
document.getElementById('compareButton').addEventListener('click', triggerComparison);
document.getElementById('reloadReports').addEventListener('click', loadReports);
document.getElementById('candidateForm').addEventListener('submit', handleCandidateSubmit);
document.getElementById('evaluationForm').addEventListener('submit', handleEvaluationSubmit);
document.getElementById('sourceForm').addEventListener('submit', handleSourceSubmit);

loadDashboard();
loadCandidates();
loadReports();
