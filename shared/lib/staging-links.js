// Staging link rewriter — included in every page after env.js
// When isStaging is true, prepends /staging to all relative links so they stay in staging scope
(function() {
    var env = window.__PAWRADISE_ENV__ || {};
    if (!env.isStaging) return;
    
    var PREFIX = '/staging';
    
    function rewrite() {
        var links = document.querySelectorAll('a[href]');
        for (var i = 0; i < links.length; i++) {
            var href = links[i].getAttribute('href');
            if (href && href.charAt(0) === '/' && href.indexOf(PREFIX) !== 0) {
                links[i].setAttribute('href', PREFIX + href);
            }
        }
        var forms = document.querySelectorAll('form[action]');
        for (var j = 0; j < forms.length; j++) {
            var action = forms[j].getAttribute('action');
            if (action && action.charAt(0) === '/' && action.indexOf(PREFIX) !== 0) {
                forms[j].setAttribute('action', PREFIX + action);
            }
        }
    }
    
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', rewrite);
    } else {
        rewrite();
    }
})();
