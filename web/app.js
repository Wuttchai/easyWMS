(() => {
  const scripts = ['/js/core.js', '/js/inventory.js', '/js/master.js', '/js/transactions.js', '/js/scanner.js'];
  let index = 0;

  function loadNext() {
    if (index === scripts.length) {
      if (token && me) {
        showApp();
        loadDashboard();
      } else {
        logout();
      }
      return;
    }
    const script = document.createElement('script');
    script.src = scripts[index++];
    script.onload = loadNext;
    script.onerror = () => console.error(`Unable to load ${script.src}`);
    document.head.appendChild(script);
  }

  loadNext();
})();
