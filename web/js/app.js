/**
 * BusTicket Pro - Frontend SPA Controller
 * Connects with Go Backend API for Fleet Admin and Interactive Booking
 */

const API_BASE = '/api';

// Application State
const state = {
  cities: [],
  busTypes: [],
  buses: [],
  routes: [],
  trips: [],
  currentTrip: null,
  selectedSeats: [], // Array of selected seat objects
  currentSeats: [],
  passengerDetails: {} // Cache for passenger input values { idx: { name, doc } }
};

// Admin Bus Builder State
const adminBuilderState = {
  mode: 'builder', // 'builder' or 'quick'
  rows: 5,
  cols: 4,
  slots: [] // Array of { row, col, active: bool, seatNum: number }
};

// ==========================================
// INITIALIZATION
// ==========================================
document.addEventListener('DOMContentLoaded', async () => {
  initTheme();
  initAdminBuilderGrid();
  await loadInitialData();
  setupEventListeners();
  const savedMod = localStorage.getItem('busticket_admin_module') || 'cities';
  switchAdminModule(savedMod);
});

async function loadInitialData() {
  await Promise.all([
    fetchCities(),
    fetchBusTypes(),
    fetchBuses(),
    fetchRoutes(),
    fetchTrips()
  ]);
  populateDropdowns();
}

function setupEventListeners() {
  // Prevent selecting identical cities in route form
  const originSelect = document.getElementById('route-origin');
  const destSelect = document.getElementById('route-destination');
  if (originSelect && destSelect) {
    originSelect.addEventListener('change', () => {
      if (originSelect.value && originSelect.value === destSelect.value) {
        showToast('Origin and Destination cities cannot be identical', 'error');
        destSelect.value = '';
      }
    });
    destSelect.addEventListener('change', () => {
      if (destSelect.value && destSelect.value === originSelect.value) {
        showToast('Origin and Destination cities cannot be identical', 'error');
        destSelect.value = '';
      }
    });
  }
}

// ==========================================
// THEME SWITCHER (DARK / LIGHT MODE)
// ==========================================
function initTheme() {
  const savedTheme = localStorage.getItem('busticket_theme') || 
    (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  applyTheme(savedTheme);
}

function toggleTheme() {
  const currentTheme = document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light';
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
  applyTheme(newTheme);
  localStorage.setItem('busticket_theme', newTheme);
  showToast(`Switched to ${newTheme === 'dark' ? 'Dark' : 'Light'} Mode`, 'info');
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  const icon = document.getElementById('theme-icon');
  const text = document.getElementById('theme-text');
  if (icon && text) {
    if (theme === 'dark') {
      icon.textContent = '☀️';
      text.textContent = 'Light Mode';
    } else {
      icon.textContent = '🌙';
      text.textContent = 'Dark Mode';
    }
  }
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

// ==========================================
// NAVIGATION & TABS
// ==========================================
function switchTab(tabName) {
  document.querySelectorAll('.tab-pane').forEach(el => el.classList.remove('active'));
  document.querySelectorAll('.nav-btn').forEach(el => el.classList.remove('active'));

  if (tabName === 'client') {
    document.getElementById('tab-client').classList.add('active');
    document.getElementById('nav-client').classList.add('active');
  } else {
    document.getElementById('tab-admin').classList.add('active');
    document.getElementById('nav-admin').classList.add('active');
    const savedMod = localStorage.getItem('busticket_admin_module') || 'cities';
    switchAdminModule(savedMod);
    loadInitialData(); // Refresh admin tables
  }
}

function switchAdminModule(moduleName) {
  // Hide all admin module panes
  document.querySelectorAll('.admin-module-pane').forEach(el => el.classList.remove('active'));
  // Deactivate all subnav buttons
  document.querySelectorAll('.admin-subnav-btn').forEach(el => el.classList.remove('active'));

  const targetPane = document.getElementById(`admin-mod-${moduleName}`);
  const targetBtn = document.getElementById(`admin-subnav-${moduleName}`);

  if (targetPane) targetPane.classList.add('active');
  if (targetBtn) targetBtn.classList.add('active');

  localStorage.setItem('busticket_admin_module', moduleName);

  // If opening fleet module, ensure the bus builder grid renders smoothly
  if (moduleName === 'fleet') {
    if (adminBuilderState.slots.length === 0) {
      initAdminBuilderGrid();
    } else {
      renderAdminBuilderGrid();
    }
  }
}

function updateAdminModuleBadges() {
  const citiesBadge = document.getElementById('badge-cities-count');
  const fleetBadge = document.getElementById('badge-fleet-count');
  const routesBadge = document.getElementById('badge-routes-count');
  const tripsBadge = document.getElementById('badge-trips-count');

  if (citiesBadge) citiesBadge.textContent = state.cities.length;
  if (fleetBadge) fleetBadge.textContent = state.buses.length;
  if (routesBadge) routesBadge.textContent = state.routes.length;
  if (tripsBadge) tripsBadge.textContent = state.trips.length;
}

// ==========================================
// TOAST NOTIFICATIONS
// ==========================================
function showToast(message, type = 'info') {
  const container = document.getElementById('toast-container');
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `
    <span>${message}</span>
    <span style="cursor:pointer; margin-left:10px; font-weight:bold;" onclick="this.parentElement.remove()">✕</span>
  `;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transition = 'opacity 0.4s ease';
    setTimeout(() => toast.remove(), 400);
  }, 4500);
}

// ==========================================
// API FETCH HELPERS
// ==========================================
async function fetchCities() {
  try {
    const res = await fetch(`${API_BASE}/admin/cities`);
    const json = await res.json();
    if (json.success) state.cities = json.data || [];
  } catch (err) {
    console.error('Error fetching cities:', err);
  }
}

async function fetchBusTypes() {
  try {
    const res = await fetch(`${API_BASE}/admin/bus-types`);
    const json = await res.json();
    if (json.success) state.busTypes = json.data || [];
  } catch (err) {
    console.error('Error fetching bus types:', err);
  }
}

async function fetchBuses() {
  try {
    const res = await fetch(`${API_BASE}/admin/buses`);
    const json = await res.json();
    if (json.success) state.buses = json.data || [];
  } catch (err) {
    console.error('Error fetching buses:', err);
  }
}

async function fetchRoutes() {
  try {
    const res = await fetch(`${API_BASE}/admin/routes`);
    const json = await res.json();
    if (json.success) state.routes = json.data || [];
  } catch (err) {
    console.error('Error fetching routes:', err);
  }
}

async function fetchTrips() {
  try {
    const res = await fetch(`${API_BASE}/admin/trips`);
    const json = await res.json();
    if (json.success) {
      state.trips = json.data || [];
      renderAdminTrips();
    }
  } catch (err) {
    console.error('Error fetching trips:', err);
  }
}

function populateDropdowns() {
  // 1. Client Search Cities
  const searchOrigin = document.getElementById('search-origin');
  const searchDest = document.getElementById('search-destination');
  const routeOrigin = document.getElementById('route-origin');
  const routeDest = document.getElementById('route-destination');

  const cityOptions = state.cities.map(c => `<option value="${c.id}">${c.nombre} (${c.terminal})</option>`).join('');
  
  if (searchOrigin) searchOrigin.innerHTML = '<option value="">Select Origin...</option>' + cityOptions;
  if (searchDest) searchDest.innerHTML = '<option value="">Select Destination...</option>' + cityOptions;
  if (routeOrigin) routeOrigin.innerHTML = '<option value="">Select Origin...</option>' + cityOptions;
  if (routeDest) routeDest.innerHTML = '<option value="">Select Destination...</option>' + cityOptions;

  // 2. Bus Types Dropdown
  const busTypeOptions = '<option value="">Select Category...</option>' + 
    state.busTypes.map(bt => `<option value="${bt.id}">${bt.nombre} (${bt.capacidad_sillas} seats)</option>`).join('');

  const busTypeSelect = document.getElementById('bus-type-select');
  if (busTypeSelect) {
    busTypeSelect.innerHTML = busTypeOptions;
  }
  const builderBusTypeSelect = document.getElementById('builder-bus-type-select');
  if (builderBusTypeSelect) {
    builderBusTypeSelect.innerHTML = busTypeOptions;
  }

  // 3. Trips Scheduler Dropdowns
  const tripRouteSelect = document.getElementById('trip-route-select');
  if (tripRouteSelect) {
    tripRouteSelect.innerHTML = '<option value="">Select Route...</option>' +
      state.routes.map(r => {
        const orig = r.ciudad_origen ? r.ciudad_origen.nombre : 'ID:' + r.ciudad_origen_id;
        const dest = r.ciudad_destino ? r.ciudad_destino.nombre : 'ID:' + r.ciudad_destino_id;
        return `<option value="${r.id}">${orig} ➔ ${dest} (${r.distancia_km} km)</option>`;
      }).join('');
  }

  const tripBusSelect = document.getElementById('trip-bus-select');
  if (tripBusSelect) {
    tripBusSelect.innerHTML = '<option value="">Select Bus...</option>' +
      state.buses.map(b => {
        const typeName = b.tipo_bus ? b.tipo_bus.nombre : '';
        return `<option value="${b.id}">Plate: ${b.placa} (Int: ${b.numero_interno}) - ${typeName}</option>`;
      }).join('');
  }

  // Render Admin Tables & Update Navigation Module Badges
  renderAdminCities();
  renderAdminBuses();
  renderAdminRoutes();
  renderAdminTrips();
  updateAdminModuleBadges();
}

// ==========================================
// CLIENT PORTAL - TRIP SEARCH & SEAT MAP
// ==========================================

async function handleSearchTrips(e) {
  e.preventDefault();
  const originId = document.getElementById('search-origin').value;
  const destId = document.getElementById('search-destination').value;
  const date = document.getElementById('search-date').value;

  if (originId === destId) {
    showToast('Origin and destination cannot be the same', 'error');
    return;
  }

  let url = `${API_BASE}/trips/search?origin_id=${originId}&destination_id=${destId}`;
  if (date) url += `&date=${encodeURIComponent(date)}`;

  try {
    const res = await fetch(url);
    const json = await res.json();

    const resultsContainer = document.getElementById('search-results-section');
    const tripsGrid = document.getElementById('trips-results-grid');

    if (!json.success || !json.data || json.data.length === 0) {
      tripsGrid.innerHTML = `<div style="grid-column: 1/-1; padding: 2rem; text-align: center; color: var(--text-muted);">
        No scheduled trips found for this route and date.
      </div>`;
      resultsContainer.style.display = 'block';
      document.getElementById('seat-booking-section').style.display = 'none';
      return;
    }

    tripsGrid.innerHTML = json.data.map(trip => {
      const departure = new Date(trip.fecha_hora_salida).toLocaleString('en-US', {
        dateStyle: 'medium',
        timeStyle: 'short'
      });
      const originName = trip.ruta && trip.ruta.ciudad_origen ? trip.ruta.ciudad_origen.nombre : 'Origin';
      const originTerm = trip.ruta && trip.ruta.ciudad_origen ? trip.ruta.ciudad_origen.terminal : '';
      const destName = trip.ruta && trip.ruta.ciudad_destino ? trip.ruta.ciudad_destino.nombre : 'Destination';
      const destTerm = trip.ruta && trip.ruta.ciudad_destino ? trip.ruta.ciudad_destino.terminal : '';
      const busPlate = trip.bus ? trip.bus.placa : 'Bus';
      const busType = trip.bus && trip.bus.tipo_bus ? trip.bus.tipo_bus.nombre : 'Standard';

      return `
        <div class="trip-card">
          <div>
            <div class="trip-header">
              <span class="trip-date">🕒 ${departure}</span>
              <span class="trip-price">$${Number(trip.precio_boleto).toLocaleString()} COP</span>
            </div>
            <div class="trip-route">
              <span>${originName}</span> ➔ <span>${destName}</span>
            </div>
            <div class="trip-terminals">
              ${originTerm} ➔ ${destTerm}
            </div>
            <div class="trip-meta">
              <span>🚌 ${busPlate} (${busType})</span>
              <span>⏱️ ${trip.ruta ? trip.ruta.duracion_estimada_minutos : '-'} min</span>
            </div>
          </div>
          <button class="btn btn-primary" onclick="selectTripForBooking(${trip.id})">
            Select & View Seats
          </button>
        </div>
      `;
    }).join('');

    resultsContainer.style.display = 'block';
    // Scroll smoothly to results
    resultsContainer.scrollIntoView({ behavior: 'smooth' });
  } catch (err) {
    showToast('Failed searching trips: ' + err.message, 'error');
  }
}

// ==========================================
// PASSENGER SELECTION & MULTI-BOOKING LOGIC
// ==========================================

function getPassengerRequirements() {
  const adultsInput = document.getElementById('search-adults');
  const childrenInput = document.getElementById('search-children');
  const adults = adultsInput ? Math.max(1, parseInt(adultsInput.value, 10) || 1) : 1;
  const children = childrenInput ? Math.max(0, parseInt(childrenInput.value, 10) || 0) : 0;
  return { adults, children, total: adults + children };
}

function handlePassengerCountChange() {
  const { adults, children, total } = getPassengerRequirements();
  
  const reqCountEl = document.getElementById('status-required-count');
  const neededBannerCount = document.getElementById('banner-needed-count');
  const passTypesBanner = document.getElementById('banner-passenger-types');
  const passBreakdown = document.getElementById('summary-passengers-breakdown');

  if (reqCountEl) reqCountEl.textContent = total;
  if (neededBannerCount) neededBannerCount.textContent = `${total} ${total === 1 ? 'seat' : 'seats'}`;
  const typesText = `${adults} ${adults === 1 ? 'Adult' : 'Adults'}${children > 0 ? `, ${children} ${children === 1 ? 'Child' : 'Children'}` : ''}`;
  if (passTypesBanner) passTypesBanner.textContent = typesText;
  if (passBreakdown) passBreakdown.textContent = typesText;

  // If user decreased count and we now have more seats selected than total, deselect excess
  if (state.selectedSeats.length > total) {
    const removedSeats = state.selectedSeats.slice(total);
    state.selectedSeats = state.selectedSeats.slice(0, total);
    removedSeats.forEach(s => {
      const btn = document.querySelector(`.seat-btn[data-seat-id="${s.silla_id}"]`);
      if (btn) {
        btn.classList.remove('selected');
        btn.classList.add('available');
      }
    });
    showToast(`Passenger count set to ${total}. Excess seats deselected.`, 'info');
  }

  updateSelectionUI();
}

async function selectTripForBooking(tripID) {
  try {
    // 1. Fetch trip details and seats
    const [tripRes, seatsRes] = await Promise.all([
      fetch(`${API_BASE}/trips/${tripID}`),
      fetch(`${API_BASE}/trips/${tripID}/seats`)
    ]);

    const tripJson = await tripRes.json();
    const seatsJson = await seatsRes.json();

    if (!tripJson.success || !seatsJson.success) {
      showToast('Could not load seat map for this trip', 'error');
      return;
    }

    state.currentTrip = tripJson.data;
    state.currentSeats = seatsJson.data;
    state.selectedSeats = [];
    state.passengerDetails = {};

    // 2. Update Checkout summary
    const trip = state.currentTrip;
    const orig = trip.ruta && trip.ruta.ciudad_origen ? trip.ruta.ciudad_origen.nombre : '';
    const dest = trip.ruta && trip.ruta.ciudad_destino ? trip.ruta.ciudad_destino.nombre : '';
    document.getElementById('summary-route').textContent = `${orig} ➔ ${dest}`;
    document.getElementById('summary-departure').textContent = new Date(trip.fecha_hora_salida).toLocaleString('en-US', {
      dateStyle: 'medium',
      timeStyle: 'short'
    });
    document.getElementById('summary-bus').textContent = `${trip.bus ? trip.bus.placa : ''} (${trip.bus && trip.bus.tipo_bus ? trip.bus.tipo_bus.nombre : ''})`;
    document.getElementById('booking-trip-id').value = trip.id;

    handlePassengerCountChange();

    // 3. Render Bus Interactive Seats Grid
    renderBusSeatGrid(state.currentSeats);

    // 4. Reveal booking section
    const bookingSec = document.getElementById('seat-booking-section');
    bookingSec.style.display = 'block';
    document.getElementById('booking-confirmation-section').style.display = 'none';
    bookingSec.scrollIntoView({ behavior: 'smooth' });

  } catch (err) {
    showToast('Error loading seat map: ' + err.message, 'error');
  }
}

function renderBusSeatGrid(seats) {
  const grid = document.getElementById('bus-seat-grid');
  grid.innerHTML = '';

  // Find max row to organize grid
  let maxRow = 1;
  seats.forEach(s => { if (s.fila > maxRow) maxRow = s.fila; });

  // Organize by rows
  for (let r = 1; r <= maxRow; r++) {
    const rowSeats = seats.filter(s => s.fila === r);
    
    // Support columns: 1, 2, [aisle], 3, 4
    for (let c = 1; c <= 4; c++) {
      if (c === 3) {
        // Central aisle
        const aisle = document.createElement('div');
        aisle.className = 'aisle-space';
        grid.appendChild(aisle);
      }

      const seat = rowSeats.find(s => s.columna === c);
      if (seat) {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'seat-btn';
        btn.textContent = seat.numero_silla;
        btn.dataset.seatId = seat.silla_id;
        btn.dataset.seatNum = seat.numero_silla;

        if (!seat.disponible) {
          btn.classList.add('occupied');
          btn.disabled = true;
          btn.title = `Seat ${seat.numero_silla}: Occupied (${seat.nombre_pasajero || 'Booked'})`;
        } else {
          // Check if currently selected
          const isSelected = state.selectedSeats.some(s => s.silla_id === seat.silla_id);
          btn.classList.add(isSelected ? 'selected' : 'available');
          btn.title = `Seat ${seat.numero_silla}: Available for selection`;
          btn.onclick = () => toggleSeatSelection(seat, btn);
        }
        grid.appendChild(btn);
      } else {
        const empty = document.createElement('div');
        grid.appendChild(empty);
      }
    }
  }
}

function toggleSeatSelection(seat, btnElement) {
  const { total } = getPassengerRequirements();
  const existingIdx = state.selectedSeats.findIndex(s => s.silla_id === seat.silla_id);

  if (existingIdx >= 0) {
    // Already selected -> DESELECT
    state.selectedSeats.splice(existingIdx, 1);
    btnElement.classList.remove('selected');
    btnElement.classList.add('available');
  } else {
    // Attempt selection
    if (state.selectedSeats.length >= total) {
      showToast(`You have already selected ${total} of ${total} required seats. Click a selected seat to deselect it.`, 'info');
      return;
    }
    state.selectedSeats.push(seat);
    btnElement.classList.remove('available');
    btnElement.classList.add('selected');
  }

  updateSelectionUI();
}

function updateSelectionUI() {
  const { adults, children, total } = getPassengerRequirements();
  const selectedCount = state.selectedSeats.length;

  const selCountEl = document.getElementById('status-selected-count');
  if (selCountEl) selCountEl.textContent = selectedCount;

  // Update chips in checkout summary
  const chipsContainer = document.getElementById('summary-seats-chips');
  if (chipsContainer) {
    if (state.selectedSeats.length === 0) {
      chipsContainer.innerHTML = '<span style="color: var(--secondary); font-size: 0.85rem;">None selected</span>';
    } else {
      chipsContainer.innerHTML = state.selectedSeats
        .map(s => `<span class="seat-badge-chip">#${s.numero_silla}</span>`)
        .join(' ');
    }
  }

  // Update total price
  const priceEl = document.getElementById('summary-price');
  if (priceEl && state.currentTrip) {
    const totalAmount = Number(state.currentTrip.precio_boleto) * selectedCount;
    priceEl.textContent = `$${totalAmount.toLocaleString()} COP`;
  }

  // Render passenger details inputs
  renderPassengerInputs(adults, children, total);

  // Enable/disable checkout button
  const bookBtn = document.getElementById('btn-book-ticket');
  if (bookBtn) {
    if (selectedCount === total && total > 0) {
      bookBtn.disabled = false;
      bookBtn.textContent = `Confirm & Book ${total} ${total === 1 ? 'Ticket' : 'Tickets'}`;
    } else {
      bookBtn.disabled = true;
      const remaining = total - selectedCount;
      bookBtn.textContent = remaining > 0 
        ? `Select ${remaining} more ${remaining === 1 ? 'seat' : 'seats'} to proceed`
        : `Confirm & Book Tickets`;
    }
  }
}

function renderPassengerInputs(adults, children, total) {
  const listContainer = document.getElementById('passengers-inputs-list');
  if (!listContainer) return;

  if (state.selectedSeats.length === 0) {
    listContainer.innerHTML = `
      <p style="color: var(--text-muted); font-size: 0.88rem; font-style: italic; margin-bottom: 1rem;">
        Please select ${total} ${total === 1 ? 'seat' : 'seats'} on the bus map above to enter passenger details.
      </p>
    `;
    return;
  }

  // Cache existing inputs
  state.selectedSeats.forEach((_, idx) => {
    const nameInput = document.getElementById(`passenger-name-${idx}`);
    const docInput = document.getElementById(`passenger-doc-${idx}`);
    if (nameInput && docInput) {
      state.passengerDetails[idx] = {
        name: nameInput.value,
        doc: docInput.value
      };
    }
  });

  listContainer.innerHTML = state.selectedSeats.map((seat, idx) => {
    const isAdult = idx < adults;
    const typeLabel = isAdult ? 'ADULT' : 'CHILD';
    const typeBadgeClass = isAdult ? 'adult' : 'child';
    const saved = state.passengerDetails[idx] || { name: '', doc: '' };

    return `
      <div class="passenger-card">
        <div class="passenger-card-header">
          <span>Passenger ${idx + 1} (Seat #${seat.numero_silla})</span>
          <span class="passenger-type-badge ${typeBadgeClass}">${typeLabel}</span>
        </div>
        <div class="form-group" style="margin-bottom: 0.6rem;">
          <label for="passenger-name-${idx}">Full Name</label>
          <input type="text" id="passenger-name-${idx}" class="form-control" 
            placeholder="e.g. ${isAdult ? 'Maria Gonzalez' : 'Lucas Gonzalez'}" 
            value="${escapeHtml(saved.name)}" required>
        </div>
        <div class="form-group">
          <label for="passenger-doc-${idx}">ID / Document Number</label>
          <input type="text" id="passenger-doc-${idx}" class="form-control" 
            placeholder="e.g. ${isAdult ? '1020304050' : 'TI-98765432'}" 
            value="${escapeHtml(saved.doc)}" required>
        </div>
      </div>
    `;
  }).join('');
}

async function handleConfirmBooking(e) {
  e.preventDefault();
  const tripId = parseInt(document.getElementById('booking-trip-id').value, 10);
  const { adults, total } = getPassengerRequirements();

  if (state.selectedSeats.length !== total) {
    showToast(`Please select all ${total} seats before confirming.`, 'error');
    return;
  }

  // Gather passenger inputs
  const passengers = [];
  for (let idx = 0; idx < state.selectedSeats.length; idx++) {
    const seat = state.selectedSeats[idx];
    const nameInput = document.getElementById(`passenger-name-${idx}`);
    const docInput = document.getElementById(`passenger-doc-${idx}`);
    const name = nameInput ? nameInput.value.trim() : '';
    const doc = docInput ? docInput.value.trim() : '';
    const isAdult = idx < adults;

    if (!name || !doc) {
      showToast(`Please fill in all details for Passenger ${idx + 1} (Seat #${seat.numero_silla})`, 'error');
      if (nameInput && !name) nameInput.focus();
      else if (docInput) docInput.focus();
      return;
    }

    passengers.push({
      silla_id: seat.silla_id,
      nombre_pasajero: name,
      documento: doc,
      tipo_pasajero: isAdult ? 'ADULT' : 'CHILD'
    });
  }

  const payload = {
    viaje_id: tripId,
    pasajeros: passengers
  };

  const btn = document.getElementById('btn-book-ticket');
  btn.disabled = true;
  btn.textContent = 'Processing reservation...';

  try {
    const res = await fetch(`${API_BASE}/tickets/book-multiple`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    const json = await res.json();

    if (res.status === 409) {
      showToast('⚠️ CONFLICT: One or more selected seats were just reserved by another customer! Please choose alternative seats.', 'error');
      await selectTripForBooking(tripId);
      return;
    }

    if (!res.ok || !json.success) {
      showToast(json.error || 'Failed to book tickets', 'error');
      btn.disabled = false;
      btn.textContent = `Confirm & Book ${total} Tickets`;
      return;
    }

    // Success! Show confirmation receipt
    showToast(`🎉 Successfully booked ${json.data.cantidad_pasajeros} tickets!`, 'success');
    renderMultiBookingReceipt(json.data);

  } catch (err) {
    showToast('Network error during booking: ' + err.message, 'error');
    btn.disabled = false;
    btn.textContent = `Confirm & Book ${total} Tickets`;
  }
}

function renderMultiBookingReceipt(multiRes) {
  document.getElementById('seat-booking-section').style.display = 'none';
  const confirmationSec = document.getElementById('booking-confirmation-section');
  const details = document.getElementById('booking-receipt-details');

  const trip = state.currentTrip;
  const origin = trip.ruta && trip.ruta.ciudad_origen ? trip.ruta.ciudad_origen.nombre : '';
  const dest = trip.ruta && trip.ruta.ciudad_destino ? trip.ruta.ciudad_destino.nombre : '';
  const departure = new Date(trip.fecha_hora_salida).toLocaleString('en-US', {
    dateStyle: 'full',
    timeStyle: 'short'
  });

  const ticketsList = multiRes.boletos.map(t => {
    const seatObj = state.currentSeats.find(s => s.silla_id === t.silla_id);
    const seatLabel = seatObj ? `Seat #${seatObj.numero_silla} (Row ${seatObj.fila}, Col ${seatObj.columna})` : `Seat ID #${t.silla_id}`;
    return `
      <div style="background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-sm); padding: 0.75rem 1rem; margin-top: 0.5rem;">
        <div style="display: flex; justify-content: space-between; font-weight: 700; margin-bottom: 0.25rem;">
          <span>Ticket #TKT-${t.id.toString().padStart(6, '0')}</span>
          <span class="badge badge-success">${t.estado}</span>
        </div>
        <div style="font-size: 0.9rem;">
          <strong>${escapeHtml(t.nombre_pasajero)}</strong> (Doc: ${escapeHtml(t.documento)}) ➔ <span style="color: var(--primary); font-weight: 600;">${seatLabel}</span>
        </div>
      </div>
    `;
  }).join('');

  details.innerHTML = `
    <p><strong>Total Tickets Issued:</strong> ${multiRes.cantidad_pasajeros}</p>
    <p><strong>Route:</strong> ${origin} ➔ ${dest}</p>
    <p><strong>Departure:</strong> ${departure}</p>
    <p><strong>Vehicle:</strong> ${trip.bus ? trip.bus.placa : ''} (${trip.bus && trip.bus.tipo_bus ? trip.bus.tipo_bus.nombre : ''})</p>
    <p><strong>Total Transaction Amount:</strong> <strong style="color: #047857; font-size: 1.15rem;">$${Number(multiRes.monto_total).toLocaleString()} COP</strong></p>
    <div style="margin-top: 1rem;">
      <h4 style="color: var(--secondary); margin-bottom: 0.5rem;">Passenger Tickets Breakdown:</h4>
      ${ticketsList}
    </div>
  `;

  confirmationSec.style.display = 'block';
  confirmationSec.scrollIntoView({ behavior: 'smooth' });
}

function resetBookingView() {
  document.getElementById('booking-confirmation-section').style.display = 'none';
  state.selectedSeats = [];
  state.passengerDetails = {};
  document.getElementById('seat-booking-section').style.display = 'none';
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

// ==========================================
// ADMIN PORTAL - CRUD OPERATIONS
// ==========================================

// --- Cities ---
async function handleCreateCity(e) {
  e.preventDefault();
  const name = document.getElementById('city-name').value.trim();
  const terminal = document.getElementById('city-terminal').value.trim();

  try {
    const res = await fetch(`${API_BASE}/admin/cities`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nombre: name, terminal: terminal })
    });
    const json = await res.json();
    if (json.success) {
      showToast(`City ${name} created successfully!`, 'success');
      document.getElementById('create-city-form').reset();
      await fetchCities();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to create city', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

async function handleDeleteCity(id) {
  if (!confirm('Are you sure you want to delete this city?')) return;
  try {
    const res = await fetch(`${API_BASE}/admin/cities/${id}`, { method: 'DELETE' });
    const json = await res.json();
    if (json.success) {
      showToast('City deleted', 'success');
      await fetchCities();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to delete city', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

function renderAdminCities() {
  const tbody = document.getElementById('cities-table-body');
  if (!tbody) return;
  tbody.innerHTML = state.cities.map(c => `
    <tr>
      <td>${c.id}</td>
      <td><strong>${c.nombre}</strong></td>
      <td>${c.terminal}</td>
      <td>
        <button class="btn btn-danger" onclick="handleDeleteCity(${c.id})">Delete</button>
      </td>
    </tr>
  `).join('');
  updateAdminModuleBadges();
}

// --- Bus Types ---
async function handleCreateBusType(e) {
  e.preventDefault();
  const name = document.getElementById('bustype-name').value.trim();
  const capacity = parseInt(document.getElementById('bustype-capacity').value, 10);

  try {
    const res = await fetch(`${API_BASE}/admin/bus-types`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nombre: name, capacidad_sillas: capacity })
    });
    const json = await res.json();
    if (json.success) {
      showToast(`Bus category ${name} (${capacity} seats) created!`, 'success');
      document.getElementById('create-bustype-form').reset();
      await fetchBusTypes();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to create bus type', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

// --- Interactive Custom Bus Builder ---
function setBusCreationMode(mode) {
  adminBuilderState.mode = mode;
  const builderPane = document.getElementById('admin-bus-builder-pane');
  const quickForm = document.getElementById('create-bus-form');
  const btnBuilder = document.getElementById('btn-mode-builder');
  const btnQuick = document.getElementById('btn-mode-quick');

  if (mode === 'builder') {
    if (builderPane) builderPane.style.display = 'block';
    if (quickForm) quickForm.style.display = 'none';
    if (btnBuilder) btnBuilder.classList.add('active');
    if (btnQuick) btnQuick.classList.remove('active');
  } else {
    if (builderPane) builderPane.style.display = 'none';
    if (quickForm) quickForm.style.display = 'block';
    if (btnBuilder) btnBuilder.classList.remove('active');
    if (btnQuick) btnQuick.classList.add('active');
  }
}

function initAdminBuilderGrid() {
  const rowsInput = document.getElementById('builder-rows');
  const colsInput = document.getElementById('builder-cols');
  const rows = rowsInput ? Math.min(15, Math.max(1, parseInt(rowsInput.value, 10) || 5)) : 5;
  const cols = colsInput ? Math.min(6, Math.max(2, parseInt(colsInput.value, 10) || 4)) : 4;

  adminBuilderState.rows = rows;
  adminBuilderState.cols = cols;
  adminBuilderState.slots = [];

  // Initialize slots. Default aisle at column 3 if 4 or 5 columns
  for (let r = 1; r <= rows; r++) {
    for (let c = 1; c <= cols; c++) {
      const isDefaultAisle = (cols === 4 && c === 3) || (cols === 5 && c === 3);
      adminBuilderState.slots.push({
        row: r,
        col: c,
        active: !isDefaultAisle,
        seatNum: 0
      });
    }
  }

  renderAdminBuilderGrid();
}

function renderAdminBuilderGrid() {
  const gridContainer = document.getElementById('admin-builder-slots-grid');
  if (!gridContainer) return;

  gridContainer.style.gridTemplateColumns = `repeat(${adminBuilderState.cols}, 48px)`;
  gridContainer.innerHTML = '';

  // Sequentially re-number active seats
  let seatSeq = 1;
  adminBuilderState.slots.forEach(slot => {
    if (slot.active) {
      slot.seatNum = seatSeq++;
    } else {
      slot.seatNum = 0;
    }
  });

  adminBuilderState.slots.forEach((slot, idx) => {
    const slotEl = document.createElement('div');
    slotEl.className = `builder-slot ${slot.active ? 'active-seat' : 'empty-aisle'}`;
    slotEl.title = slot.active 
      ? `Seat #${slot.seatNum} (Row ${slot.row}, Col ${slot.col}) - Click to toggle Aisle` 
      : `Aisle Space (Row ${slot.row}, Col ${slot.col}) - Click to toggle Seat`;
    slotEl.innerHTML = slot.active ? `<span>#${slot.seatNum}</span>` : `<span style="font-size:0.75rem;">Aisle</span>`;
    slotEl.onclick = () => toggleBuilderSlot(idx);
    gridContainer.appendChild(slotEl);
  });

  const activeCount = adminBuilderState.slots.filter(s => s.active).length;
  const countEl = document.getElementById('builder-active-seats-count');
  if (countEl) countEl.textContent = activeCount;

  const dimEl = document.getElementById('builder-dimensions-label');
  if (dimEl) dimEl.textContent = `${adminBuilderState.rows} Rows x ${adminBuilderState.cols} Cols`;
}

function toggleBuilderSlot(index) {
  if (index >= 0 && index < adminBuilderState.slots.length) {
    adminBuilderState.slots[index].active = !adminBuilderState.slots[index].active;
    renderAdminBuilderGrid();
  }
}

async function handleSaveCustomBus(e) {
  e.preventDefault();
  const plate = document.getElementById('builder-bus-plate').value.trim();
  const typeId = parseInt(document.getElementById('builder-bus-type-select').value, 10);
  const internalNum = document.getElementById('builder-bus-internal-num').value.trim();

  const activeSlots = adminBuilderState.slots.filter(s => s.active);
  if (activeSlots.length === 0) {
    showToast('Please enable at least 1 seat in the visual bus designer', 'error');
    return;
  }

  const customSeats = activeSlots.map(s => ({
    numero_silla: s.seatNum,
    fila: s.row,
    columna: s.col
  }));

  const payload = {
    placa: plate,
    tipo_bus_id: typeId,
    numero_interno: internalNum,
    sillas: customSeats
  };

  const saveBtn = document.getElementById('btn-save-custom-bus');
  if (saveBtn) {
    saveBtn.disabled = true;
    saveBtn.textContent = 'Saving custom bus...';
  }

  try {
    const res = await fetch(`${API_BASE}/admin/buses`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (json.success) {
      showToast(`🎉 Custom bus ${plate} created with ${customSeats.length} seats!`, 'success');
      document.getElementById('custom-bus-builder-form').reset();
      initAdminBuilderGrid();
      await fetchBuses();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to create custom bus', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  } finally {
    if (saveBtn) {
      saveBtn.disabled = false;
      saveBtn.textContent = '💾 Save Bus with Custom Seats';
    }
  }
}

// --- Quick Auto-Generate Buses ---
async function handleCreateBus(e) {
  e.preventDefault();
  const plate = document.getElementById('bus-plate').value.trim();
  const typeId = parseInt(document.getElementById('bus-type-select').value, 10);
  const internalNum = document.getElementById('bus-internal-num').value.trim();

  try {
    const res = await fetch(`${API_BASE}/admin/buses`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ placa: plate, tipo_bus_id: typeId, numero_interno: internalNum })
    });
    const json = await res.json();
    if (json.success) {
      const seatCount = json.data && json.data.sillas ? json.data.sillas.length : 0;
      showToast(`Bus ${plate} added and ${seatCount} seats auto-generated in grid!`, 'success');
      document.getElementById('create-bus-form').reset();
      await fetchBuses();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to create bus', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

async function handleDeleteBus(id) {
  if (!confirm('Are you sure you want to delete this bus?')) return;
  try {
    const res = await fetch(`${API_BASE}/admin/buses/${id}`, { method: 'DELETE' });
    const json = await res.json();
    if (json.success) {
      showToast('Bus removed', 'success');
      await fetchBuses();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to delete bus', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

function renderAdminBuses() {
  const tbody = document.getElementById('buses-table-body');
  if (!tbody) return;
  tbody.innerHTML = state.buses.map(b => {
    const typeName = b.tipo_bus ? b.tipo_bus.nombre : '-';
    const capacity = b.tipo_bus ? b.tipo_bus.capacidad_sillas : '-';
    return `
      <tr>
        <td>${b.id}</td>
        <td><strong>${b.placa}</strong></td>
        <td><span class="badge badge-info">${typeName}</span></td>
        <td>#${b.numero_interno}</td>
        <td>${capacity} seats</td>
        <td>
          <button class="btn btn-danger" onclick="handleDeleteBus(${b.id})">Delete</button>
        </td>
      </tr>
    `;
  }).join('');
  updateAdminModuleBadges();
}

// --- Routes ---
async function handleCreateRoute(e) {
  e.preventDefault();
  const originId = parseInt(document.getElementById('route-origin').value, 10);
  const destId = parseInt(document.getElementById('route-destination').value, 10);
  const distance = parseFloat(document.getElementById('route-distance').value);
  const duration = parseInt(document.getElementById('route-duration').value, 10);

  if (originId === destId) {
    showToast('Origin and Destination cities cannot be identical', 'error');
    return;
  }

  try {
    const res = await fetch(`${API_BASE}/admin/routes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ciudad_origen_id: originId,
        ciudad_destino_id: destId,
        distancia_km: distance,
        duracion_estimada_minutos: duration
      })
    });
    const json = await res.json();
    if (json.success) {
      showToast('Route added successfully!', 'success');
      document.getElementById('create-route-form').reset();
      await fetchRoutes();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to create route', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

async function handleDeleteRoute(id) {
  if (!confirm('Are you sure you want to delete this route?')) return;
  try {
    const res = await fetch(`${API_BASE}/admin/routes/${id}`, { method: 'DELETE' });
    const json = await res.json();
    if (json.success) {
      showToast('Route deleted', 'success');
      await fetchRoutes();
      populateDropdowns();
    } else {
      showToast(json.error || 'Failed to delete route', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

function renderAdminRoutes() {
  const tbody = document.getElementById('routes-table-body');
  if (!tbody) return;
  tbody.innerHTML = state.routes.map(r => {
    const orig = r.ciudad_origen ? r.ciudad_origen.nombre : 'ID: ' + r.ciudad_origen_id;
    const dest = r.ciudad_destino ? r.ciudad_destino.nombre : 'ID: ' + r.ciudad_destino_id;
    return `
      <tr>
        <td>${r.id}</td>
        <td><strong>${orig}</strong></td>
        <td><strong>${dest}</strong></td>
        <td>${r.distancia_km} km</td>
        <td>${r.duracion_estimada_minutos} min</td>
        <td>
          <button class="btn btn-danger" onclick="handleDeleteRoute(${r.id})">Delete</button>
        </td>
      </tr>
    `;
  }).join('');
  updateAdminModuleBadges();
}

// --- Trips Scheduler ---
async function handleCreateTrip(e) {
  e.preventDefault();
  const routeId = parseInt(document.getElementById('trip-route-select').value, 10);
  const busId = parseInt(document.getElementById('trip-bus-select').value, 10);
  const departureInput = document.getElementById('trip-departure-time').value;
  const price = parseFloat(document.getElementById('trip-price').value);

  if (!departureInput) {
    showToast('Please select a departure date and time', 'error');
    return;
  }

  // Convert to ISO string
  const departureDate = new Date(departureInput).toISOString();

  try {
    const res = await fetch(`${API_BASE}/admin/trips`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ruta_id: routeId,
        bus_id: busId,
        fecha_hora_salida: departureDate,
        precio_boleto: price
      })
    });
    const json = await res.json();
    if (json.success) {
      showToast('Trip scheduled successfully!', 'success');
      document.getElementById('create-trip-form').reset();
      await fetchTrips();
    } else {
      showToast(json.error || 'Failed to schedule trip', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

async function handleDeleteTrip(id) {
  if (!confirm('Are you sure you want to delete this scheduled trip?')) return;
  try {
    const res = await fetch(`${API_BASE}/admin/trips/${id}`, { method: 'DELETE' });
    const json = await res.json();
    if (json.success) {
      showToast('Trip deleted', 'success');
      await fetchTrips();
    } else {
      showToast(json.error || 'Failed to delete trip', 'error');
    }
  } catch (err) {
    showToast('Error: ' + err.message, 'error');
  }
}

function renderAdminTrips() {
  const tbody = document.getElementById('trips-table-body');
  if (!tbody) return;
  tbody.innerHTML = state.trips.map(v => {
    const orig = v.ruta && v.ruta.ciudad_origen ? v.ruta.ciudad_origen.nombre : '';
    const dest = v.ruta && v.ruta.ciudad_destino ? v.ruta.ciudad_destino.nombre : '';
    const plate = v.bus ? v.bus.placa : '';
    const departure = new Date(v.fecha_hora_salida).toLocaleString('en-US', {
      dateStyle: 'medium',
      timeStyle: 'short'
    });
    return `
      <tr>
        <td>${v.id}</td>
        <td>${orig} ➔ ${dest}</td>
        <td><strong>${plate}</strong></td>
        <td>${departure}</td>
        <td>$${Number(v.precio_boleto).toLocaleString()}</td>
        <td>
          <button class="btn btn-danger" onclick="handleDeleteTrip(${v.id})">Delete</button>
        </td>
      </tr>
    `;
  }).join('');
  updateAdminModuleBadges();
}
