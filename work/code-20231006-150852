function TIC()
	local t=time()
	for x=-120,119 do
		for y=-68,67 do
			local p=1.5+math.sin(t/1000)*.4
   local a=math.atan2(y,x)/(math.pi*2)+(p*t)/10000
			local d=(x^2+y^2)^.5
   local c=1+(a*2+d^p/100)%2
			pix(120+x,68+y,c)
		end
	end 
end
