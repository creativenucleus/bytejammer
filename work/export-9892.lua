-- pos: 19,3
function TIC()t=time()//32
for y=0,136 do for x=0,240 do
pix(x,y,(x+y+t)>>4)
end end end
