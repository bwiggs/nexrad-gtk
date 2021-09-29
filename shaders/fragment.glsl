#version 330

// It was expressed that some drivers required this next line
// to function properly
precision highp float;

in vec3 f_color;

void main(void) {
    // Pass through original color with full opacity.
    gl_FragColor = vec4(f_color,1.0);
}