const MAX_DIM = 2048;

export async function resizeImage(blob: Blob): Promise<string> {
	const bmp = await createImageBitmap(blob);
	let { width, height } = bmp;

	const longest = Math.max(width, height);
	if (longest > MAX_DIM) {
		const scale = MAX_DIM / longest;
		width = Math.round(width * scale);
		height = Math.round(height * scale);
	}

	const canvas = document.createElement('canvas');
	canvas.width = width;
	canvas.height = height;
	const ctx = canvas.getContext('2d')!;
	ctx.drawImage(bmp, 0, 0, width, height);
	bmp.close();

	const dataUrl = canvas.toDataURL('image/png');
	return dataUrl.replace(/^data:[^;]+;base64,/, '');
}
