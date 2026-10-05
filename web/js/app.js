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
  selectedSeat: null,
  currentSeats: []
};

// ==========================================
// INITIALIZATION
// ==========================================
document.addEventListener('DOMContentLoaded', async () => {
  await loadInitialData();
  setupEventListeners();
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
    loadInitialData(); // Refresh admin tables
  }
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
  const busTypeSelect = document.getElementById('bus-type-select');
  if (busTypeSelect) {
    busTypeSelect.innerHTML = '<option value="">Select Category...</option>' + 
      state.busTypes.map(bt => `<option value="${bt.id}">${bt.nombre} (${bt.capacidad_sillas} seats)</option>`).join('');
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

  // Render Admin Tables
  renderAdminCities();
  renderAdminBuses();
  renderAdminRoutes();
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
    state.selectedSeat = null;

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
    document.getElementById('summary-seat').textContent = 'None selected';
    document.getElementById('summary-price').textContent = `$${Number(trip.precio_boleto).toLocaleString()} COP`;

    document.getElementById('booking-trip-id').value = trip.id;
    document.getElementById('booking-seat-id').value = '';
    document.getElementById('btn-book-ticket').disabled = true;

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
    
    // Sort by column: 1, 2, [aisle], 3, 4
    for (let c = 1; c <= 4; c++) {
      if (c === 3) {
        // Insert aisle placeholder
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
          // OCCUPIED (Red, disabled)
          btn.classList.add('occupied');
          btn.disabled = true;
          btn.title = `Seat ${seat.numero_silla}: Occupied (${seat.nombre_pasajero || 'Booked'})`;
        } else {
          // AVAILABLE (Green, clickable)
          btn.classList.add('available');
          btn.title = `Seat ${seat.numero_silla}: Available for selection`;
          btn.onclick = () => selectSeat(seat, btn);
        }
        grid.appendChild(btn);
      } else {
        // Empty slot
        const empty = document.createElement('div');
        grid.appendChild(empty);
      }
    }
  }
}

function selectSeat(seat, btnElement) {
  // Clear previous selection
  document.querySelectorAll('.seat-btn.selected').forEach(el => {
    el.classList.remove('selected');
    el.classList.add('available');
  });

  // Apply selection
  btnElement.classList.remove('available');
  btnElement.classList.add('selected');

  state.selectedSeat = seat;
  document.getElementById('booking-seat-id').value = seat.silla_id;
  document.getElementById('summary-seat').textContent = `Seat #${seat.numero_silla} (Row ${seat.fila}, Col ${seat.columna})`;
  document.getElementById('btn-book-ticket').disabled = false;
}

async function handleConfirmBooking(e) {
  e.preventDefault();
  const tripId = parseInt(document.getElementById('booking-trip-id').value, 10);
  const seatId = parseInt(document.getElementById('booking-seat-id').value, 10);
  const name = document.getElementById('passenger-name').value.trim();
  const doc = document.getElementById('passenger-doc').value.trim();
  const status = document.getElementById('payment-status').value;

  if (!tripId || !seatId) {
    showToast('Please select an available seat first', 'error');
    return;
  }

  const payload = {
    viaje_id: tripId,
    silla_id: seatId,
    nombre_pasajero: name,
    documento: doc,
    estado: status
  };

  const btn = document.getElementById('btn-book-ticket');
  btn.disabled = true;
  btn.textContent = 'Processing reservation...';

  try {
    const res = await fetch(`${API_BASE}/tickets/book`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    const json = await res.json();

    if (res.status === 409) {
      // RACE CONDITION DETECTED!
      showToast('⚠️ CONFLICT: This seat was just reserved by another customer at the same time! Please choose another seat.', 'error');
      // Refresh seat map immediately to show the seat occupied in red
      await selectTripForBooking(tripId);
      return;
    }

    if (!res.ok || !json.success) {
      showToast(json.error || 'Failed to book ticket', 'error');
      btn.disabled = false;
      btn.textContent = 'Confirm & Reserve Seat';
      return;
    }

    // Success! Show confirmation receipt
    showToast('🎉 Ticket successfully booked!', 'success');
    renderBookingReceipt(json.data);

  } catch (err) {
    showToast('Network error during booking: ' + err.message, 'error');
    btn.disabled = false;
    btn.textContent = 'Confirm & Reserve Seat';
  }
}

function renderBookingReceipt(ticket) {
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

  details.innerHTML = `
    <p><strong>Ticket ID / Code:</strong> #TKT-${ticket.id.toString().padStart(6, '0')}</p>
    <p><strong>Passenger:</strong> ${ticket.nombre_pasajero} (Doc: ${ticket.documento})</p>
    <p><strong>Route:</strong> ${origin} ➔ ${dest}</p>
    <p><strong>Seat:</strong> Seat #${state.selectedSeat.numero_silla} (Row ${state.selectedSeat.fila}, Col ${state.selectedSeat.columna})</p>
    <p><strong>Departure:</strong> ${departure}</p>
    <p><strong>Vehicle:</strong> ${trip.bus ? trip.bus.placa : ''} (${trip.bus && trip.bus.tipo_bus ? trip.bus.tipo_bus.nombre : ''})</p>
    <p><strong>Status:</strong> <span class="badge badge-success">${ticket.estado}</span></p>
    <p><strong>Total Amount:</strong> $${Number(trip.precio_boleto).toLocaleString()} COP</p>
  `;

  confirmationSec.style.display = 'block';
  confirmationSec.scrollIntoView({ behavior: 'smooth' });
}

function resetBookingView() {
  document.getElementById('booking-confirmation-section').style.display = 'none';
  document.getElementById('passenger-name').value = '';
  document.getElementById('passenger-doc').value = '';
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

// --- Buses ---
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
}
