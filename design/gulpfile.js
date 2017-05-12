var gulp = require('gulp')
var sass = require('gulp-sass')
var autoprefixer = require('gulp-autoprefixer')
var rename = require('gulp-rename');
var minifyCSS = require('gulp-minify-css');
var concat = require('gulp-concat');

gulp.task('sass', function () {
  return gulp.src('static/css/_scss/style.scss')
    .pipe(sass()) // Using gulp-sass
    .pipe(gulp.dest('static/css'))
})

gulp.task('autoprefixer', function () {
  return gulp.src('static/css/style.css')
    .pipe(autoprefixer({
      browsers: ['last 2 versions'],
      cascade: false
    }))
    .pipe(gulp.dest('static/css'))
})

gulp.task('minifyCSS', function() {
  return gulp.src('static/css/style.css')
    .pipe(minifyCSS())
    .pipe(rename('style.min.css'))
    .pipe(gulp.dest('static/css'))
})

gulp.task('watch', function () {
  gulp.watch('static/css/_scss/*', ['sass'])
  gulp.watch('static/css/*', ['autoprefixer'])
  gulp.watch('static/css/*', ['minifyCSS'])
})