async function postAction(url,body,msgId){const m=$(msgId);try{const d=await api(url,{method:'POST',body:JSON.stringify(body)});m.className='msg success';m.textContent=(d.message||'completed')+' ✓';loadDashboard();return d}catch(e){m.className='msg error';m.textContent=e.message;throw e}}
async function submitMovement(type){if(type==='issue')return submitIssue();if(type==='receive'&&$('receiveSubmit').disabled)return;const p=type==='receive'?'r':'i',body={sku:$(p+'Sku').value,location_code:$(p+'Loc').value,qty:Number($(p+'Qty').value),lot_no:$(p+'Lot').value,reference:$(p+'Ref').value,reason_code:$(p+'Reason').value};if(type==='receive'){body.supplier_code=$('rSupplier').value;if($('rMfg').value)body.mfg_date=new Date($('rMfg').value+'T00:00:00Z').toISOString();if($('rExp').value)body.exp_date=new Date($('rExp').value+'T00:00:00Z').toISOString()}if(type==='receive')$('receiveSubmit').disabled=true;try{await postAction('/api/'+type,body,type+'Msg');if(type==='receive'){clearReceiveForm();showReceiveHistory('รับสินค้าเข้าคลังเรียบร้อยแล้ว');loadLocations()}}catch(e){alert(e.message)}finally{if(type==='receive')$('receiveSubmit').disabled=false}}
async function submitTransfer() {
  if ($('transferSubmit').disabled) return;
  if (!$('tFrom').value) {
    $('transferMsg').className = 'msg error';
    $('transferMsg').textContent = 'เลือกต้นทางที่มีสินค้าใน Inventory';
    $('transferMsg').textContent = validationMessage($('transferMsg').textContent);
    alert($('transferMsg').textContent);
    $('tFrom').focus();
    return;
  }
  $('transferSubmit').disabled = true;
  try {
    await postAction('/api/transfer',{sku:$('tSku').value,from_location:$('tFrom').value,to_location:$('tTo').value,qty:Number($('tQty').value),reference:$('tRef').value,reason_code:$('tReason').value},'transferMsg');
    await showTransferHistory('ย้ายสินค้าเรียบร้อยแล้ว');
    await loadLocations();
  } catch (e) { alert(e.message); } finally { $('transferSubmit').disabled = false; }
}

async function openTransferForm() {
  if ($('transferSubmit').disabled) return;
  ['tSku','tFrom','tTo','tQty','tRef'].forEach(id => $(id).value = '');
  transferSkuAuto = false;
  $('tProductPicker').classList.add('hidden');
  $('tReason').value = 'TRF';
  $('transferMsg').textContent = '';
  $('transferMsg').className = 'msg';
  $('transferHistory').classList.add('hidden');
  $('transferForm').classList.remove('hidden');
  $('pageSub').textContent = 'ย้ายสินค้าระหว่าง Location';
  await Promise.all([loadLocations(), loadTransferSources()]);
  $('tFrom').focus();
}

let transferInventory = [], transferSourceLoadId = 0;
let transferSkuAuto = false;

async function loadTransferSources() {
  const loadId = ++transferSourceLoadId;
  transferInventory = [];
  $('tFrom').disabled = true;
  $('tFrom').innerHTML = '<option value="">กำลังโหลด Inventory...</option>';
  try {
    const rows = await api('/api/inventory');
    if (loadId !== transferSourceLoadId) return;
    transferInventory = (rows || []).filter(row => Number(row.qty) > 0 && row.location?.code);
    updateTransferSources();
  } catch (e) {
    if (loadId !== transferSourceLoadId) return;
    $('tFrom').innerHTML = '<option value="">โหลด Inventory ไม่สำเร็จ</option>';
    $('transferMsg').className = 'msg error';
    $('transferMsg').textContent = e.message;
  }
}

function updateTransferSources() {
  const selected = $('tFrom').value;
  const sku = transferSkuAuto ? '' : $('tSku').value.trim();
  const rows = transferInventory.filter(row => !sku || row.product?.sku === sku || row.product?.barcode === sku);
  const locations = [...new Map(rows.map(row => [row.location.code, row.location])).values()];
  $('tFrom').innerHTML = `<option value="">${locations.length ? 'เลือกต้นทางจาก Inventory' : 'ไม่พบสต็อกคงเหลือ'}</option>` + locations.map(location => {
    const stock = rows.filter(row => row.location.code === location.code).map(row => `${row.product?.sku || '-'}: ${fmt(row.qty)} ${row.product?.unit || ''}`).join(', ');
    return `<option value="${esc(location.code)}">${esc(location.code)} - ${esc(stock)}</option>`;
  }).join('');
  $('tFrom').value = locations.some(location => location.code === selected) ? selected : '';
  $('tFrom').disabled = !locations.length;
  updateTransferProducts(false);
}

function editTransferSku() {
  transferSkuAuto = false;
  updateTransferSources();
}

function updateTransferProducts(autoFill = true) {
  const rows = transferInventory.filter(row => row.location.code === $('tFrom').value && row.product?.sku);
  const products = [...new Map(rows.map(row => [row.product.sku, row.product])).values()];
  const current = $('tSku').value.trim();
  const matching = products.find(product => product.sku === current || product.barcode === current);
  $('tProduct').innerHTML = '<option value="">เลือกสินค้า</option>' + products.map(product => `<option value="${esc(product.sku)}">${esc(product.sku)} - ${esc(product.name)}</option>`).join('');
  $('tProductPicker').classList.toggle('hidden', products.length <= 1);
  $('tProduct').value = matching?.sku || '';
  if (autoFill) {
    $('tSku').value = products.length === 1 ? products[0].sku : matching?.sku || '';
    $('tProduct').value = $('tSku').value;
    transferSkuAuto = true;
    updateTransferSources();
  }
}

function selectTransferProduct() {
  $('tSku').value = $('tProduct').value;
  transferSkuAuto = true;
  updateTransferSources();
}

function showTransferHistory(message = '') {
  $('transferForm').classList.add('hidden');
  $('transferHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการย้ายสินค้า';
  $('transferHistoryMsg').className = 'msg success';
  $('transferHistoryMsg').textContent = message;
  return loadTransferHistory();
}

const transferPageSize = 5;
let transferPage = 1, transferRows = [], transferHistoryLoadId = 0;

async function loadTransferHistory() {
  const loadId = ++transferHistoryLoadId;
  transferPage = 1;
  transferRows = [];
  $('transferTable').textContent = 'กำลังโหลดประวัติการย้ายสินค้า...';
  try {
    // Each transfer has two movements; show its outgoing movement once.
    const rows = await api('/api/movements?type=TRANSFER_OUT');
    if (loadId !== transferHistoryLoadId) return;
    transferRows = rows || [];
    renderTransferPage();
  } catch (e) {
    if (loadId !== transferHistoryLoadId) return;
    $('transferTable').textContent = 'ไม่สามารถโหลดประวัติการย้ายสินค้า: ' + e.message;
  }
}

function renderTransferPage() {
  renderSearchTable('transferTable', table(['วันที่ย้าย','เอกสารอ้างอิง','SKU','สินค้า','ต้นทาง','ปลายทาง','จำนวน','หน่วย','ผู้ย้าย'], transferRows.map(x => `<tr><td data-search-date="${dateSearchValue(x.created_at)}">${dateFmt(x.created_at)}</td><td>${esc(x.reference || '-')}</td><td>${esc(x.product?.sku || '-')}</td><td>${esc(x.product?.name || '-')}</td><td>${esc(x.location?.code || '-')}</td><td>${esc(x.note?.startsWith('To ') ? x.note.slice(3) : '-')}</td><td>${fmt(x.qty)}</td><td>${esc(x.product?.unit || '-')}</td><td>${esc(x.created_by || '-')}</td></tr>`)), {page: transferPage, pageSize: transferPageSize, onPage: page => { transferPage = page; }});
}

function changeTransferPage(direction) {
  transferPage += direction;
  renderTransferPage();
}
async function submitAdjustment() {
  if ($('adjustmentSubmit').disabled) return;
  $('adjustmentSubmit').disabled = true;
  try {
    await postAction('/api/adjustments',{sku:$('aSku').value,location_code:$('aLoc').value,direction:$('aDirection').value,qty:Number($('aQty').value),reference:$('aRef').value,reason_code:$('aReason').value,note:$('aNote').value},'adjustmentMsg');
    await showAdjustmentHistory('ส่งคำขอปรับสต็อกแล้ว รออนุมัติ');
  } catch (e) { alert(e.message); } finally { $('adjustmentSubmit').disabled = false; }
}

async function openAdjustmentForm() {
  if ($('adjustmentSubmit').disabled) return;
  ['aSku','aLoc','aQty','aRef','aNote','aReason'].forEach(id => $(id).value = '');
  $('aDirection').value = 'IN';
  $('adjustmentMsg').textContent = '';
  $('adjustmentMsg').className = 'msg';
  $('adjustmentHistory').classList.add('hidden');
  $('adjustmentForm').classList.remove('hidden');
  $('pageSub').textContent = 'สร้างคำขอปรับสต็อก';
  await loadLocations();
  $('aSku').focus();
}

function showAdjustmentHistory(message = '') {
  $('adjustmentForm').classList.add('hidden');
  $('adjustmentHistory').classList.remove('hidden');
  $('pageSub').textContent = 'อนุมัติ/ปฏิเสธการปรับ Stock';
  $('adjustmentHistoryMsg').className = 'msg success';
  $('adjustmentHistoryMsg').textContent = message;
  return loadAdjustments();
}

const adjustmentPageSize = 5;
let adjustmentRows = [], adjustmentPage = 1, adjustmentLoadId = 0;

async function loadAdjustments(page = 1) {
  const loadId = ++adjustmentLoadId;
  adjustmentRows = [];
  adjustmentPage = page;
  $('adjustmentTable').textContent = 'กำลังโหลดคำขอปรับสต็อก...';
  try {
    const rows = await api('/api/adjustments');
    if (loadId !== adjustmentLoadId) return;
    adjustmentRows = rows || [];
    renderAdjustmentPage();
  } catch (e) {
    if (loadId !== adjustmentLoadId) return;
    $('adjustmentTable').textContent = 'โหลดคำขอไม่สำเร็จ: ' + e.message;
  }
}

function renderAdjustmentPage() {
  const rows = adjustmentRows.map(x => {
    const can = me && ['ADMIN','SUPERVISOR'].includes(me.role) && x.status === 'PENDING';
    return `<tr><td data-search-date="${dateSearchValue(x.created_at)}">${dateFmt(x.created_at)}</td><td>${esc(x.document_no)}</td><td>${esc(x.product?.sku)}</td><td>${esc(x.location?.code)}</td><td>${badge(x.direction)}</td><td>${fmt(x.qty)}</td><td>${esc(x.reason_code || '-')}</td><td>${badge(x.status)}</td><td>${esc(x.requested_by || '-')}</td><td>${can ? `<button class="mini" onclick="adjAction('${esc(x.id)}','approve')">Approve</button> <button class="mini danger" onclick="adjAction('${esc(x.id)}','reject')">Reject</button>` : '-'}</td></tr>`;
  });
  renderSearchTable('adjustmentTable', table(['วันที่','เอกสาร','SKU','Location','ประเภท','จำนวน','เหตุผล','สถานะ','ผู้ขอ','ดำเนินการ'], rows), {page: adjustmentPage, pageSize: adjustmentPageSize, onPage: page => { adjustmentPage = page; }});
}

function changeAdjustmentPage(direction) {
  adjustmentPage += direction;
  renderAdjustmentPage();
}
function clearReceiveForm() {
  $('rProduct').value = '';
  $('rSupplier').value = '';
  ['rSku','rLoc','rQty','rLot','rMfg','rExp','rRef'].forEach(id => $(id).value = '');
  $('rReason').value = 'RCV';
  $('receiveMsg').textContent = '';
  $('receiveMsg').className = 'msg';
}

async function openReceiveForm() {
  clearReceiveForm();
  $('receiveHistory').classList.add('hidden');
  $('receiveForm').classList.remove('hidden');
  $('pageSub').textContent = 'รับสินค้าเข้าคลัง';
  await Promise.all([loadLocations(), loadReceiveSuppliers(), loadReceiveProducts()]);
  ($('rProduct').disabled ? $('rSku') : $('rProduct')).focus();
}

let receiveProducts = [], receiveProductsLoadId = 0;

async function loadReceiveProducts() {
  const loadId = ++receiveProductsLoadId;
  receiveProducts = [];
  $('rProduct').disabled = true;
  $('rProduct').innerHTML = '<option value="">เลือกสินค้า</option>';
  $('receiveProductHint').textContent = 'กำลังโหลดสินค้า...';
  try {
    const products = await api('/api/products');
    if (loadId !== receiveProductsLoadId) return;
    receiveProducts = products || [];
    $('rProduct').innerHTML += receiveProducts.map(product => `<option value="${esc(product.sku)}">${esc(product.sku)} - ${esc(product.name)}${product.unit ? ' (' + esc(product.unit) + ')' : ''}</option>`).join('');
    $('rProduct').disabled = !receiveProducts.length;
    $('receiveProductHint').textContent = receiveProducts.length ? '' : 'ยังไม่มีสินค้า เพิ่มได้ที่ Master Data > Product';
    syncReceiveProduct();
  } catch (e) {
    if (loadId !== receiveProductsLoadId) return;
    $('receiveProductHint').textContent = 'โหลดสินค้าไม่สำเร็จ: ' + e.message;
  }
}

function selectReceiveProduct() {
  $('rSku').value = $('rProduct').value;
}

function syncReceiveProduct() {
  const value = $('rSku').value.trim();
  const product = receiveProducts.find(product => product.sku === value || (product.barcode && product.barcode === value));
  $('rProduct').value = product?.sku || '';
}

async function loadReceiveSuppliers() {
  $('rSupplier').innerHTML = '<option value="">ไม่ระบุผู้ขาย</option>';
  $('rSupplier').disabled = true;
  $('receiveSupplierHint').textContent = 'กำลังโหลดผู้ขาย...';
  try {
    const suppliers = await api('/api/suppliers');
    $('rSupplier').innerHTML += (suppliers || []).map(x => `<option value="${esc(x.code)}">${esc(x.code)} - ${esc(x.name)}</option>`).join('');
    $('rSupplier').disabled = false;
    $('receiveSupplierHint').textContent = suppliers?.length ? '' : 'ยังไม่มีผู้ขาย เพิ่มได้ที่ Master Data > Supplier';
  } catch (e) {
    $('receiveSupplierHint').textContent = 'โหลดผู้ขายไม่สำเร็จ: ' + e.message + ' กรุณาปิดแล้วเปิดฟอร์มใหม่';
  }
}

function showReceiveHistory(message = '') {
  $('receiveForm').classList.add('hidden');
  $('receiveHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการรับสินค้า';
  $('receiveHistoryMsg').className = 'msg success';
  $('receiveHistoryMsg').textContent = message;
  return loadReceiveHistory();
}

const receivePageSize = 5;
let receivePage = 1, receiveRows = [], receiveLoadId = 0;
const receiveHeaders = ['วันที่รับ','เอกสารอ้างอิง','ผู้ขาย','SKU','สินค้า','คลังสินค้า','Location','จำนวน','หน่วย','Lot / Batch','ผู้รับ'];

function receiveCells(x) {
  return [dateFmt(x.created_at), x.reference || '-', x.supplier_code ? x.supplier_code + ' - ' + x.supplier_name : '-', x.product?.sku || '-', x.product?.name || '-', x.location?.warehouse_code || '-', x.location?.code || '-', fmt(x.qty), x.product?.unit || '-', x.lot_no || '-', x.created_by || '-'];
}



async function loadReceiveHistory() {
  const loadId = ++receiveLoadId;
  const target = $('receiveTable');
  receivePage = 1;
  receiveRows = [];
  target.textContent = 'กำลังโหลดประวัติการรับสินค้า...';
  try {
    const rows = await api('/api/movements?type=IN');
    if (loadId !== receiveLoadId) return;
    receiveRows = rows || [];
    renderReceivePage();
  } catch (e) {
    if (loadId !== receiveLoadId) return;
    target.textContent = 'ไม่สามารถโหลดประวัติการรับสินค้า: ' + e.message;
  }
}

function renderReceivePage() {
  renderSearchTable('receiveTable', table(receiveHeaders, receiveRows.map(x =>
    `<tr>${receiveCells(x).map((value, index) => `<td${index === 0 ? ` data-search-date="${dateSearchValue(x.created_at)}"` : ''}>${esc(value)}</td>`).join('')}</tr>`
  )), {page: receivePage, pageSize: receivePageSize, onPage: page => { receivePage = page; }});
}

function changeReceivePage(direction) {
  receivePage += direction;
  renderReceivePage();
}
let issueInventory = [], issueLots = [], issueLoading = false, issueSaving = false, issueLoadId = 0;

async function loadIssueInventory() {
  const loadId = ++issueLoadId;
  issueLoading = true;
  updateIssueAvailable();
  try {
    const [inventory, lots] = await Promise.all([api('/api/inventory'), api('/api/inventory-lots')]);
    if (loadId !== issueLoadId) return false;
    issueInventory = inventory || [];
    issueLots = lots || [];
    issueLoading = false;
    updateIssueLocations();
    return true;
  } catch (e) {
    if (loadId !== issueLoadId) return false;
    issueInventory = [];
    issueLots = [];
    issueLoading = false;
    updateIssueLocations();
    $('issueMsg').className = 'msg error';
    $('issueMsg').textContent = e.message;
    return false;
  }
}

let issueChoices = [], issueChoiceKey = '';

function updateIssueLocations() {
  const field = $('iLoc'), selected = field.value;
  const locations = [...new Map(issueInventory.filter(x => Number(x.qty) > 0).map(x => [x.location.code, x.location])).values()].sort((a,b) => a.code.localeCompare(b.code));
  field.innerHTML = '<option value="">เลือก Location ที่ต้องการเบิก</option>' + locations.map(x => `<option value="${esc(x.code)}">${esc(x.code)} — ${esc(x.name)} (${esc(x.warehouse_code)})</option>`).join('');
  field.value = locations.some(x => x.code === selected) ? selected : locations.length === 1 ? locations[0].code : '';
  field.disabled = !locations.length || issueLoading || issueSaving;
  updateIssueItems();
}

function updateIssueItems() {
  issueChoices = [];
  for (const stock of issueInventory.filter(x => x.location.code === $('iLoc').value && Number(x.qty) > 0)) {
    const lots = issueLots.filter(x => x.product_id === stock.product_id && x.location_id === stock.location_id && Number(x.qty) > 0).sort((a,b) => (a.exp_date || '9999').localeCompare(b.exp_date || '9999'));
    for (const lot of lots) issueChoices.push({key: lot.id, stock, lot: lot.lot_no, expiry: lot.exp_date, qty: Math.min(Number(stock.qty),Number(lot.qty))});
    const untracked = Number(stock.qty) - lots.reduce((sum,x) => sum + Number(x.qty),0);
    if (untracked > 0) issueChoices.push({key: stock.id, stock, lot: '', expiry: null, qty: untracked});
  }
  if (!issueChoices.some(x => x.key === issueChoiceKey)) issueChoiceKey = issueChoices.length === 1 ? issueChoices[0].key : '';
  $('issueItems').innerHTML = !$('iLoc').value ? '<p>เลือก Location เพื่อดูสินค้าที่เบิกได้</p>' : table(['เลือก','สินค้า','Lot','หมดอายุ','คงเหลือ'], issueChoices.map((x,i) => `<tr><td><input type="radio" name="issueChoice" aria-label="${esc(x.stock.product.sku + ' ' + (x.lot || 'ไม่ระบุ Lot'))}" ${x.key === issueChoiceKey ? 'checked' : ''} onchange="selectIssueItem(${i})" style="width:auto"></td><td>${esc(x.stock.product.sku)}<br>${esc(x.stock.product.name)}</td><td>${esc(x.lot || 'ไม่มี Lot')}</td><td data-search-date="${dateSearchValue(x.expiry)}">${dFmt(x.expiry)}</td><td>${fmt(x.qty)} ${esc(x.stock.product.unit)}</td></tr>`));
  updateIssueAvailable();
}

function selectIssueItem(index) {
  issueChoiceKey = issueChoices[index]?.key || '';
  $('iQty').value = '1';
  updateIssueAvailable();
}

function selectedIssueItem() { return issueChoices.find(x => x.key === issueChoiceKey); }

function updateIssueAvailable() {
  const item = selectedIssueItem(), qty = Number($('iQty').value);
  const valid = item && Number.isFinite(qty) && qty > 0 && qty <= item.qty;
  $('iQty').max = String(item?.qty || 0);
  $('iQty').disabled = !item || issueLoading || issueSaving;
  $('iLoc').disabled = issueLoading || issueSaving || !issueInventory.some(x => Number(x.qty)>0);
  document.querySelectorAll('[name="issueChoice"]').forEach(x => x.disabled = issueLoading || issueSaving);
  $('issueSubmit').disabled = issueLoading || issueSaving;
  $('issueSubmit').textContent = issueSaving ? 'กำลังบันทึก...' : issueLoading ? 'กำลังโหลดสต็อก...' : 'ยืนยันเบิกสินค้า';
  $('issueAvailable').textContent = issueLoading ? 'กำลังโหลด Inventory...' : !issueInventory.some(x => Number(x.qty)>0) ? 'ไม่มีสินค้าคงเหลือให้เบิก' : !item ? 'เลือก Location และรายการสินค้าเพื่อเบิก' : `เบิกได้สูงสุด ${fmt(item.qty)} ${item.stock.product.unit}${valid ? '' : ' — ระบุจำนวนมากกว่า 0 และไม่เกินยอดคงเหลือ'}`;
  $('issueSummary').textContent = item ? `${item.stock.product.name} • ${item.stock.location.code} • Lot: ${item.lot || '-'} • จำนวน ${fmt(qty)} ${item.stock.product.unit}` : '';
  if (item) $('issueSummary').textContent += $('iPurpose').value === 'CUSTOMER' ? ' • ลูกค้า: ' + ($('iCustomer').value ? $('iCustomer').selectedOptions[0].textContent : 'ยังไม่ได้เลือก') : ' • เบิกใช้ภายใน';
}

async function submitIssue() {
  if (issueSaving || issueLoading) return;
  const item = selectedIssueItem(), qty = Number($('iQty').value);
  let error = '', field = $('iLoc');
  if (!issueInventory.some(x => Number(x.qty) > 0)) error = 'ไม่มีสินค้าคงเหลือให้เบิก กรุณา Refresh เพื่อตรวจสอบสต็อกอีกครั้ง';
  else if (!$('iLoc').value) error = 'กรุณาเลือก Location ที่ต้องการเบิกสินค้า';
  else if (!item) {
    error = 'กรุณาเลือกรายการสินค้า / Lot ที่ต้องการเบิก';
    field = document.querySelector('[name="issueChoice"]') || $('iLoc');
  } else if (!$('iQty').value.trim() || !Number.isFinite(qty) || qty <= 0) {
    error = 'กรุณาระบุจำนวนที่ต้องการเบิกเป็นตัวเลขมากกว่า 0';
    field = $('iQty');
  } else if (qty > item.qty) {
    error = `จำนวนเบิกเกินยอดคงเหลือ เบิกได้สูงสุด ${fmt(item.qty)} ${item.stock.product.unit}`;
    field = $('iQty');
  }
  if (!error && $('iPurpose').value === 'CUSTOMER' && !$('iCustomer').value) {
    error = 'กรุณาเลือกลูกค้าสำหรับการเบิกส่งลูกค้า';
    field = $('iCustomer');
  }
  if (error) { error = validationMessage(error);
    $('issueMsg').className = 'msg error';
    $('issueMsg').textContent = error;
    alert(error);
    field.focus();
    return;
  }
  $('issueMsg').textContent = '';
  if (!confirm('ยืนยันเบิกสินค้า?\n' + $('issueSummary').textContent)) return;
  const body = {sku: item.stock.product.sku, location_code: item.stock.location.code, qty, lot_no: item.lot, reference: $('iRef').value.trim(), reason_code: 'ISS', note: $('iNote').value.trim()};
  body.issue_purpose = $('iPurpose').value;
  body.customer_code = body.issue_purpose === 'CUSTOMER' ? $('iCustomer').value : '';
  issueSaving = true;
  updateIssueAvailable();
  try {
    await postAction('/api/issue', body, 'issueMsg');
    await showIssueHistory('เบิกสินค้าเรียบร้อยแล้ว');
    $('iQty').value = '1';
    await loadIssueInventory();
  } catch (e) {
    alert('ไม่สามารถเบิกสินค้าได้: ' + e.message);
    await loadIssueInventory();
  } finally {
    issueSaving = false;
    updateIssueAvailable();
  }
}

const issuePageSize = 5;
let issuePage = 1, issueRows = [], issueHistoryLoadId = 0;

async function loadIssueHistory() {
  const loadId = ++issueHistoryLoadId;
  const target = $('issueTable');
  issuePage = 1;
  issueRows = [];
  target.textContent = 'กำลังโหลดประวัติการเบิกสินค้า...';
  try {
    const rows = await api('/api/movements?type=OUT');
    if (loadId !== issueHistoryLoadId) return;
    issueRows = rows || [];
    renderIssuePage();
  } catch (e) {
    if (loadId !== issueHistoryLoadId) return;
    target.textContent = 'ไม่สามารถโหลดประวัติการเบิกสินค้า: ' + e.message;
  }
}

function renderIssuePage() {
  renderSearchTable('issueTable', table(
      ['วันที่เบิก','เอกสารอ้างอิง','ลูกค้า','SKU','สินค้า','คลังสินค้า','Location','จำนวน','หน่วย','Lot / Batch','ผู้เบิก'],
      issueRows.map(x => `<tr><td data-search-date="${dateSearchValue(x.created_at)}">${dateFmt(x.created_at)}</td><td>${esc(x.reference || '-')}</td><td>${esc(x.customer_code ? x.customer_code + ' - ' + x.customer_name : x.issue_purpose === 'INTERNAL' ? 'เบิกใช้ภายใน' : '-')}</td><td>${esc(x.product?.sku || '-')}</td><td>${esc(x.product?.name || '-')}</td><td>${esc(x.location?.warehouse_code || '-')}</td><td>${esc(x.location?.code || '-')}</td><td>${fmt(x.qty)}</td><td>${esc(x.product?.unit || '-')}</td><td>${esc(x.lot_no || '-')}</td><td>${esc(x.created_by || '-')}</td></tr>`)
    ), {page: issuePage, pageSize: issuePageSize, onPage: page => { issuePage = page; }});
}

function changeIssuePage(direction) {
  issuePage += direction;
  renderIssuePage();
}
async function openIssueForm() {
  if (issueSaving) return;
  issueChoiceKey = '';
  issueChoices = [];
  $('iLoc').value = '';
  $('iQty').value = '1';
  $('iRef').value = '';
  $('iNote').value = '';
  $('iPurpose').value = 'INTERNAL';
  $('iCustomer').value = '';
  updateIssueCustomer();
  $('issueMsg').textContent = '';
  $('issueMsg').className = 'msg';
  $('issueForm').querySelector('details').open = false;
  $('issueHistory').classList.add('hidden');
  $('issueForm').classList.remove('hidden');
  $('pageSub').textContent = 'เบิกสินค้าออก';
  await Promise.all([loadIssueInventory(), loadIssueCustomers()]);
  $('iLoc').focus();
}

function updateIssueCustomer() {
  const required = $('iPurpose').value === 'CUSTOMER';
  $('issueCustomerField').classList.toggle('hidden', !required);
  $('iCustomer').required = required;
  if (!required) $('iCustomer').value = '';
  updateIssueAvailable();
}

async function loadIssueCustomers() {
  $('iCustomer').innerHTML = '<option value="">เลือกลูกค้า</option>';
  $('iCustomer').disabled = true;
  $('issueCustomerHint').textContent = 'กำลังโหลดลูกค้า...';
  try {
    const customers = await api('/api/customers');
    $('iCustomer').innerHTML += (customers || []).map(x => `<option value="${esc(x.code)}">${esc(x.code)} - ${esc(x.name)}</option>`).join('');
    $('iCustomer').disabled = false;
    $('issueCustomerHint').textContent = customers?.length ? '' : 'ยังไม่มีลูกค้า กรุณาเพิ่มที่ Master Data > Customer';
  } catch (e) {
    $('issueCustomerHint').textContent = 'โหลดลูกค้าไม่สำเร็จ: ' + e.message + ' กรุณาปิดแล้วเปิดฟอร์มใหม่';
  }
}

function showIssueHistory(message = '') {
  $('issueForm').classList.add('hidden');
  $('issueHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการเบิกสินค้า';
  $('issueHistoryMsg').className = 'msg success';
  $('issueHistoryMsg').textContent = message;
  return loadIssueHistory();
}
let countInventory = [], countLocations = [], countProducts = [], countLoading = false, countSaving = false, countLoadId = 0;

async function loadCountInventory() {
  const loadId = ++countLoadId;
  countLoading = true;
  updateCountPreview();
  try {
    const [inventory, locations, products] = await Promise.all([api('/api/inventory'), api('/api/locations'), api('/api/products')]);
    if (loadId !== countLoadId) return;
    countInventory = inventory || [];
    countLocations = locations || [];
    countProducts = products || [];
  } catch (e) {
    if (loadId !== countLoadId) return;
    countInventory = []; countLocations = []; countProducts = [];
    $('stockCountMsg').className = 'msg error';
    $('stockCountMsg').textContent = 'โหลดข้อมูลตรวจนับไม่สำเร็จ: ' + e.message;
  }
  if (loadId !== countLoadId) return;
  countLoading = false;
  renderCountLocations();
}

function renderCountLocations() {
  const selected = $('cLoc').value;
  const found = $('cMode').value === 'found';
  const ids = new Set(countInventory.map(x => x.location_id));
  const locations = countLocations.filter(x => found || ids.has(x.id));
  $('cLoc').innerHTML = '<option value="">เลือก Location</option>' + locations.map(x => `<option value="${esc(x.code)}">${esc(x.code)} - ${esc(x.name)} (${esc(x.warehouse_code)})</option>`).join('');
  $('cLoc').value = locations.some(x => x.code === selected) ? selected : '';
  renderCountProducts();
}

function countProductOptions() {
  const location = countLocations.find(x => x.code === $('cLoc').value);
  if (!location) return [];
  const ids = new Set(countInventory.filter(x => x.location_id === location.id).map(x => x.product_id));
  return countProducts.filter(x => $('cMode').value === 'found' ? !ids.has(x.id) : ids.has(x.id));
}

function renderCountProducts() {
  const selected = $('cSku').value, products = countProductOptions();
  $('cSku').innerHTML = '<option value="">เลือกสินค้า</option>' + products.map(x => `<option value="${esc(x.sku)}">${esc(x.sku)} - ${esc(x.name)}</option>`).join('');
  $('cSku').value = products.some(x => x.sku === selected) ? selected : products.length === 1 ? products[0].sku : '';
  resetCountQuantity();
}

function resetCountQuantity() {
  $('cQty').value = '';
  updateCountPreview();
}

function countSelection() {
  const product = countProductOptions().find(x => x.sku === $('cSku').value);
  const location = countLocations.find(x => x.code === $('cLoc').value);
  if (!product || !location) return null;
  const stock = countInventory.find(x => x.product_id === product.id && x.location_id === location.id);
  return {product, location, qty: Number(stock?.qty || 0)};
}

function updateCountPreview() {
  const selection = countSelection(), raw = $('cQty').value, qty = Number(raw);
  const busy = countLoading || countSaving;
  $('cMode').disabled = busy;
  $('cLoc').disabled = busy || $('cLoc').options.length <= 1;
  $('cSku').disabled = busy || !countProductOptions().length;
  $('cQty').disabled = busy || !selection;
  $('countSubmit').disabled = busy;
  $('countSubmit').textContent = countLoading ? 'กำลังโหลด...' : countSaving ? 'กำลังบันทึก...' : 'บันทึกผลตรวจนับ';
  $('countSystem').textContent = selection ? `ยอดในระบบ ${fmt(selection.qty)} ${selection.product.unit}` : !$('cLoc').value ? 'เลือก Location เพื่อเริ่มตรวจนับ (รวมรายการที่ยอดเป็น 0)' : 'ไม่มีสินค้าที่เลือกได้ในรูปแบบนี้ หรือยังไม่ได้เลือกสินค้า';
  $('countDifference').textContent = !selection || !raw.trim() ? '' : !Number.isFinite(qty) || qty < 0 ? 'กรุณาระบุจำนวนตั้งแต่ 0 ขึ้นไป' : `ผลต่าง ${fmt(qty - selection.qty)} ${selection.product.unit} — ${qty === selection.qty ? 'ยอดตรงกับระบบ' : 'ต้องขออนุมัติปรับยอดก่อนเปลี่ยนสต็อก'}`;
}

async function submitStockCount() {
  if (countLoading || countSaving) return;
  const selection = countSelection(), raw = $('cQty').value, qty = Number(raw);
  let error = '', field = $('cLoc');
  if (!$('cLoc').value) error = 'กรุณาเลือก Location ที่ต้องการตรวจนับ';
  else if (!selection) { error = 'กรุณาเลือกสินค้าที่ต้องการตรวจนับ'; field = $('cSku'); }
  else if (!raw.trim() || !Number.isFinite(qty) || qty < 0) { error = 'กรุณากรอกจำนวนที่นับได้จริงตั้งแต่ 0 ขึ้นไป'; field = $('cQty'); }
  if (error) { error = validationMessage(error); $('stockCountMsg').className = 'msg error'; $('stockCountMsg').textContent = error; alert(error); field.focus(); return; }
  if (!confirm(`ยืนยันผลตรวจนับ ${selection.product.name} ที่ ${selection.location.code}\nยอดในระบบ ${fmt(selection.qty)} / นับได้ ${fmt(qty)} / ผลต่าง ${fmt(qty-selection.qty)}\n${qty === selection.qty ? 'ยอดตรงกับระบบ' : 'ผลต่างจะส่งขออนุมัติปรับยอด'}`)) return;
  countSaving = true;
  updateCountPreview();
  try {
    const result = await postAction('/api/stock-count', {sku:selection.product.sku, location_code:selection.location.code, counted_qty:qty, reference:$('cRef').value.trim(), reason_code:$('cReason').value.trim()}, 'stockCountMsg');
    await showCountHistory(result.data?.status === 'PENDING_APPROVAL' ? 'บันทึกผลตรวจนับแล้ว รออนุมัติปรับยอด' : 'บันทึกผลตรวจนับแล้ว ยอดตรงกับระบบ');
    $('cRef').value = '';
  } catch (e) { alert('ไม่สามารถบันทึกผลตรวจนับได้: ' + e.message); }
  finally { countSaving = false; updateCountPreview(); }
}

async function openCountForm() {
  if (countSaving) return;
  $('cMode').value = 'inventory';
  ['cLoc','cSku','cQty','cRef'].forEach(id => $(id).value = '');
  $('cReason').value = 'CNT';
  $('stockCountMsg').textContent = '';
  $('stockCountMsg').className = 'msg';
  $('countHistory').classList.add('hidden');
  $('countForm').classList.remove('hidden');
  $('pageSub').textContent = 'ตรวจนับสินค้า';
  await loadCountInventory();
  $('cLoc').focus();
}

function showCountHistory(message = '') {
  $('countForm').classList.add('hidden');
  $('countHistory').classList.remove('hidden');
  $('pageSub').textContent = 'ประวัติการตรวจนับสินค้า';
  $('countHistoryMsg').className = 'msg success';
  $('countHistoryMsg').textContent = message;
  return loadStockCounts();
}

const countHistoryPageSize = 5;
let countHistoryRows = [], countHistoryPage = 1, countHistoryLoadId = 0;

async function loadStockCounts() {
  const loadId = ++countHistoryLoadId;
  countHistoryRows = [];
  countHistoryPage = 1;
  $('stockCountTable').textContent = 'กำลังโหลดประวัติการตรวจนับ...';
  try {
    const rows = await api('/api/stock-counts');
    if (loadId !== countHistoryLoadId) return;
    countHistoryRows = rows || [];
    renderCountHistory();
  } catch (e) {
    if (loadId !== countHistoryLoadId) return;
    $('stockCountTable').textContent = 'ไม่สามารถโหลดประวัติการตรวจนับ: ' + e.message;
  }
}

function renderCountHistory() {
  renderSearchTable('stockCountTable', table(['วันที่นับ','เอกสาร','SKU','สินค้า','Location','ยอดในระบบ','นับได้จริง','ผลต่าง','สถานะ','ผู้ตรวจนับ'], countHistoryRows.map(x => `<tr><td data-search-date="${dateSearchValue(x.created_at)}">${dateFmt(x.created_at)}</td><td>${esc(x.document_no)}</td><td>${esc(x.product?.sku)}</td><td>${esc(x.product?.name)}</td><td>${esc(x.location?.code)}</td><td>${fmt(x.system_qty)}</td><td>${fmt(x.counted_qty)}</td><td>${fmt(x.difference)}</td><td>${badge(x.status)}</td><td>${esc(x.created_by)}</td></tr>`)), {page: countHistoryPage, pageSize: countHistoryPageSize, onPage: page => { countHistoryPage = page; }});
}

function changeCountHistoryPage(direction) {
  countHistoryPage += direction;
  renderCountHistory();
}
