var gulp = require('gulp')
var sass = require('gulp-sass')
var autoprefixer = require('gulp-autoprefixer')
var rename = require('gulp-rename');
var minifyCSS = require('gulp-minify-css');
var concat = require('gulp-concat');

gulp.task('sass', function () {
  return gulp.src('new_app/static/css/_scss/style.scss')
    .pipe(sass()) // Using gulp-sass
    .pipe(gulp.dest('new_app/static/css'))
})

gulp.task('autoprefixer', function () {
  return gulp.src('new_app/static/css/style.css')
    .pipe(autoprefixer({
      browsers: ['last 2 versions'],
      cascade: false
    }))
    .pipe(gulp.dest('new_app/static/css'))
})

gulp.task('minifyCSS', function() {
  return gulp.src('new_app/static/css/style.css')
    .pipe(minifyCSS())
    .pipe(rename('style.min.css'))
    .pipe(gulp.dest('new_app/static/css'))
})

gulp.task('watch', function () {
  gulp.watch('new_app/static/css/_scss/*', ['sass'])
  gulp.watch('new_app/static/css/*', ['autoprefixer'])
  gulp.watch('new_app/static/css/*', ['minifyCSS'])
})