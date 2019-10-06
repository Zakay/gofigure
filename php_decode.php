<?php

while (false !== ($line = fgets(STDIN))) {
  var_dump(json_decode($line, true));
}