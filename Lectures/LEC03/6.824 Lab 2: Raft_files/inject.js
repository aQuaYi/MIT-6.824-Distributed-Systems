(function inject() {
  let style;
  let mouseRightButtonDown;
  let paused = false;

  const events = [
    'cut',
    'copy',
    'mousedown',
    'touchstart',
    'selectstart',
    'contextmenu',
  ];

  const css = '* { -moz-user-select: text !important; user-select: text !important }';

  function setEventProperty(object, property) {
    const descriptor = Object.getOwnPropertyDescriptor(object, property);
    Object.defineProperty(object, property, {
      configurable: true,
      enumerable: true,
      get() {
        if (descriptor && descriptor.get) {
          return descriptor.get.call(this);
        }
        return null;
      },
      set(v) {
        let value;
        if (typeof v === 'function') {
          value = (event) => {
            const returnValue = v.call(this, event);
            if (paused
            || (event.type === 'click' && event.button === 0 && !event.ctrlKey && !event.shiftKey)) {
              return returnValue;
            }
            return undefined;
          };
        } else {
          value = v;
        }
        if (descriptor && descriptor.set) {
          descriptor.set.call(this, value);
        }
      },
    });
  }

  function hackInlineEvent(el, eventProperty) {
    const fn = el[eventProperty];
    el.removeAttribute(eventProperty);
    el[eventProperty] = fn;
  }

  function checkInlineEvent(eventProperty, event) {
    if (paused || event.target === document) {
      return;
    }
    if (event.target.hasAttribute(eventProperty)) {
      hackInlineEvent(event.target, eventProperty);
    }
    let el = event.target.closest(`[${eventProperty}]`);
    while (el) {
      hackInlineEvent(el, eventProperty);
      el = el.closest(`[${eventProperty}]`);
    }
  }

  window.addEventListener('message', (event) => {
    if (event.source !== window || event.data.msg !== 'ercc:pause-script') {
      return;
    }
    paused = event.data.value;
    if (paused) {
      style.remove();
    } else {
      document.head.appendChild(style);
      style.sheet.insertRule(css, 0);
    }
  });

  const original = {
    preventDefault: Event.prototype.preventDefault,
    alert: window.alert,
    confirm: window.confirm,
    prompt: window.prompt,
  };

  Object.defineProperty(Event.prototype, 'preventDefault', {
    configurable: true,
    writable: false,
    enumerable: true,
    value() {
      if (paused || !events.includes(this.type)
      || (this.type === 'click' && this.button === 0 && !this.ctrlKey && !this.shiftKey)) {
        original.preventDefault.call(this);
      }
    },
  });

  let eventType;
  for (eventType of events) {
    const eventProperty = `on${eventType}`;
    setEventProperty(window, eventProperty);
    setEventProperty(Document.prototype, eventProperty);
    setEventProperty(HTMLElement.prototype, eventProperty);
    window.addEventListener(eventType, checkInlineEvent.bind(null, eventProperty), true);
  }

  window.addEventListener('mousedown', (e) => {
    if (paused || e.button !== 2) {
      return;
    }
    mouseRightButtonDown = true;
  }, true);

  window.addEventListener('mouseup', (e) => {
    if (paused || e.button !== 2) {
      return;
    }
    setTimeout(() => { mouseRightButtonDown = false; }, 0);
  }, true);

  ['alert', 'confirm', 'prompt'].forEach((prop) => {
    window[prop] = function fn(...args) {
      if (paused || !mouseRightButtonDown) {
        original[prop].apply(this, args);
      }
    };
  });

  style = document.createElement('style');
  document.head.appendChild(style);
  style.sheet.insertRule(css, 0);
}());
