function copy_function() {
    // Get the URL from the button's data-url attribute
    const url = document.getElementById('copyButton').getAttribute('data-url');

    if (!url) {
        alert('Error: URL not found!');
        return;
    }

    // Create a temporary input element to hold the full URL
    const tempInput = document.createElement('input');
    tempInput.value = "http://localhost:8000/" + url; // Prefix with your domain
    document.body.appendChild(tempInput);

    // Select the text inside the temporary input element
    tempInput.select();
    tempInput.setSelectionRange(0, 99999); // For mobile devices

    // Execute the copy command
    try {
        const successful = document.execCommand('copy');
        if (successful) {
            alert('URL copied to clipboard!');
        } else {
            alert('Failed to copy URL. Please try manually.');
        }
    } catch (err) {
        console.error('Copy failed:', err);
        alert('Error: Unable to copy URL.');
    }

    // Remove the temporary input element
    document.body.removeChild(tempInput);
}