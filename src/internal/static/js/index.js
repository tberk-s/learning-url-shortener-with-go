document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('shortenForm');
    const resultDiv = document.getElementById('result');
    const loading = document.getElementById('loading');
    const errorDiv = document.getElementById('error');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        // Reset states
        errorDiv.style.display = 'none';
        loading.style.display = 'block';
        resultDiv.innerHTML = '';

        const url = document.getElementById('urlInput').value;

        try {
            const response = await fetch('/shorten', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: `url=${encodeURIComponent(url)}`
            });

            const responseText = await response.text();

            if (!response.ok) {
                errorDiv.textContent = responseText;
                errorDiv.style.display = 'block';
                return;
            }

            // Replace the entire page content with the success template
            document.documentElement.innerHTML = responseText;
        } catch (error) {
            errorDiv.textContent = 'Network error occurred. Please try again.';
            errorDiv.style.display = 'block';
        } finally {
            loading.style.display = 'none';
        }
    });
});
